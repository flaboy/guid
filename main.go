package main

import (
	// "bufio"
	// "bytes"
	"container/list"
	"flag"
	"fmt"
	"gopkg.in/redis.v5"
	// "io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var (
	redis_server    *string
	redis_password  *string
	redis_key       *string
	idlen           *uint
	watermark_low   *int
	generate_number *int
	redis_client    *redis.Client
	jump_number     *int
	command         string
	rnd             *rand.Rand
	step_id         int
)

const (
	interval = 1
)

func main() {

	if len(os.Args) > 1 {
		command = os.Args[1]
		os.Args = os.Args[1:]
	}

	parse_arg(command)
	flag.Parse()

	switch command {
	case "start":
		log.Printf("redis=%s, idlen=%d, key=\"%s\"\n", *redis_server, *idlen, *redis_key)
		do_start_server()
	case "top":
		do_top()
	case "clear-redis":
		do_clear_redis()
	default:
		topic := flag.Arg(1)
		if topic != "" {
			fmt.Fprintf(os.Stderr, "%s %s <options>%s\noptions:\n", os.Args[0], topic, command_arg_line_info(topic))
			parse_arg(topic)
			flag.PrintDefaults()
		} else {
			print_help()
		}
		if command != "help" {
			os.Exit(1)
		}
	}
}

func print_help() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "   start       - start service")
	fmt.Fprintln(os.Stderr, "   top         - get top 10 id in redis")
	fmt.Fprintln(os.Stderr, "   clear-redis - truncate id list in redis")
	fmt.Fprintln(os.Stderr, "\nMore: guid help <command>")
}

func command_arg_line_info(command string) string {
	switch command {
	case "has":
		return " <test>"
	}
	return ""
}

// functions....

func redis_conn() {
	redis_client = redis.NewClient(&redis.Options{
		Addr:     *redis_server,
		Password: *redis_password,
		DB:       0,
	})
}

func watchloop() {
	for {
		llen := redis_client.LLen(*redis_key)
		if llen.Err() == nil {
			if int(llen.Val()) < *watermark_low {
				log.Printf("count(\"%s\")=%d < %d, generate ids\n", *redis_key, llen.Val(), *watermark_low)
				generate_id_list(*redis_key, *idlen)
			}
		} else {
			log.Println("redis-error:", llen.Err())
			time.Sleep(time.Second * interval)
		}
	}
}

func generate_id_list(key string, idlen uint) (err error) {

	var (
		i      = int(math.Pow(10, float64(idlen-2)))
		max    = int(math.Pow(10, float64(idlen-1)))
		el     *list.Element
		cnt    = 0
		prefix = strconv.Itoa(step_id)
	)

	l := list.New()
	mp := make([]*list.Element, max)

	for i < max {
		i += rnd.Intn(*jump_number)
		if l.Len() == 0 {
			el = l.PushFront(i)
		} else {
			el = l.InsertBefore(i, mp[rnd.Intn(cnt)])
		}
		mp[cnt] = el
		cnt += 1
	}

	el = l.Front()
	for el != nil {
		redis_client.RPush(key, prefix+strconv.Itoa(el.Value.(int)))
		el = el.Next()
	}
	step_id += 1
	return nil
}

func parse_arg(command string) {
	idlen = flag.Uint("l", 7, "id length.")

	if command == "start" {
		watermark_low = flag.Int("m", 50000, "list length watermark.")
		generate_number = flag.Int("n", 100000, "id numbers per generate action.")
		jump_number = flag.Int("j", 10, "jump number")
	}

	switch command {
	case "top", "clear-redis", "start":
		redis_server = flag.String("s", "127.0.0.1:6379", "redis server address")
		redis_password = flag.String("p", "", "redis password")
		redis_key = flag.String("k", "guid-"+strconv.Itoa(int(*idlen)), "redis id-key")
	}
}

// command....

func do_start_server() {
	redis_conn()
	step_id = 0
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

	_, err := redis_client.Ping().Result()
	if err == nil {
		log.Printf("redis connected, starting watchloop for \"%s\"\n", *redis_key)
		watchloop()
	} else {
		log.Fatal("redis", err)
	}
}

func do_top() {
	redis_conn()
	rst, err := redis_client.LRange(*redis_key, 0, 10).Result()
	if err == nil {
		for _, id := range rst {
			fmt.Println(id)
		}
	} else {
		fmt.Println(err)
		os.Exit(1)
	}
}

func do_clear_redis() {
	redis_conn()
	err := redis_client.Del(*redis_key).Err()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
