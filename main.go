package main

import (
	"container/list"
	"flag"
	"fmt"
	"gopkg.in/redis.v5"
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
	redis_key       string
	idlen           *uint
	watermark_low   *int
	generate_number *int
	redis_client    *redis.Client
	jump_number     *int
	command         string
	rnd             *rand.Rand
	step_id         int64
	prefix          *string
)

const (
	interval = 1
)

func main() {

	prog_name := os.Args[0]
	if len(os.Args) > 1 {
		command = os.Args[1]
		os.Args = os.Args[1:]
	}

	parse_arg(command)
	flag.Parse()

	switch command {
	case "start":
		get_redis_key(prog_name, command)
		log.Printf("redis=%s, idlen=%d, key=\"%s\"\n", *redis_server, *idlen, redis_key)
		do_start_server()
	case "top":
		get_redis_key(prog_name, command)
		do_top()
	case "clear-redis":
		get_redis_key(prog_name, command)
		do_clear_redis()
	default:
		if len(os.Args) > 1 {
			topic := os.Args[1]
			fmt.Fprintf(os.Stderr, "%s %s <options>%s\noptions:\n", prog_name, topic, command_arg_line_info(topic))
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

func get_redis_key(prog_name, command string) {
	redis_key = flag.Arg(0)
	if redis_key == "" {
		fmt.Fprintf(os.Stderr, "empty redis-key name, usage: %s %s <options> <redis-key>\n", prog_name, command)
		os.Exit(1)
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
	case "start":
		return " <redis-key>"
	case "top":
		return " <redis-key>"
	case "clear-redis":
		return " <redis-key>"
	}
	return ""
}

func redis_conn() {
	redis_client = redis.NewClient(&redis.Options{
		Addr:     *redis_server,
		Password: *redis_password,
		DB:       0,
	})
}

func watchloop() {
	for {
		llen := redis_client.LLen(redis_key)
		if llen.Err() == nil {
			if int(llen.Val()) < *watermark_low {
				log.Printf("count(\"%s\")=%d < %d, generate ids\n", redis_key, llen.Val(), *watermark_low)
				generate_id_list(redis_key, *idlen)
			}
		} else {
			log.Println("redis-error:", llen.Err())
		}
		time.Sleep(time.Second * interval)
	}
}

func generate_id_list(key string, idlen uint) (err error) {

	step_id = redis_client.Incr(key + "_step").Val()

	var (
		i          = int(math.Pow(10, float64(idlen-1)))
		max        = int(math.Pow(10, float64(idlen)))
		el         *list.Element
		cnt        = 0
		sub_prefix = strconv.FormatInt(step_id, 10)
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
		redis_client.RPush(key, *prefix+sub_prefix+strconv.Itoa(el.Value.(int)))
		el = el.Next()
	}
	return nil
}

func parse_arg(command string) {

	if command == "start" {
		watermark_low = flag.Int("m", 50000, "list length watermark.")
		jump_number = flag.Int("j", 10, "jump number")
		idlen = flag.Uint("l", 6, "id length.")
		prefix = flag.String("a", "", "prefix")
	}

	switch command {
	case "top", "clear-redis", "start":
		redis_server = flag.String("s", "127.0.0.1:6379", "redis server address")
		redis_password = flag.String("p", "", "redis password")
	}
}

func do_start_server() {
	redis_conn()
	step_id = 0
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

	_, err := redis_client.Ping().Result()
	if err == nil {
		log.Printf("redis connected, starting watchloop for \"%s\"\n", redis_key)
		step_id_str := redis_client.Get(redis_key + "_step").Val()
		step_id, err := strconv.ParseInt(step_id_str, 10, 0)
		if err != nil {
			redis_client.Set(redis_key+"_step", 10000, 0)
			log.Printf("set step=10000\n")
		} else {
			log.Printf("get step=%d\n", step_id)
		}
		watchloop()
	} else {
		log.Fatal("redis", err)
	}
}

func do_top() {
	redis_conn()
	rst, err := redis_client.LRange(redis_key, 0, 10).Result()
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
	err := redis_client.Del(redis_key).Err()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
