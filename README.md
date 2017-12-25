guid
======================

Generate uniq-id and RPUSH redis list.

Install:
---------------------

```
go get github.com/flaboy/guid
```

Commands:
---------------------

```
Usage of help:
Commands:
   start       - start service
   top         - get top 10 id in redis
   clear-redis - truncate id list in redis

More: guid help <command>

./guid start <options> <redis-key>
options:
  -a string
    	prefix
  -j int
    	jump number (default 10)
  -l uint
    	id length. (default 6)
  -m int
    	list length watermark. (default 50000)
  -p string
    	redis password
  -s string
    	redis server address (default "127.0.0.1:6379")
```
