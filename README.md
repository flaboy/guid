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
   import      - import id list to bloomfilter from a file
   top         - get top 10 id in redis
   clear-redis - truncate id list in redis
   has         - test id in bloomfilter

More: guid help <command>
