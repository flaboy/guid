Guid
======================

Generate uniq-id and RPUSH redis list.

Install:
---------------------

```
go get github.com/flaboy/guid
```

Example:

start service
```
guid start guid-order
017/12/25 18:09:17 redis=127.0.0.1:6379, idlen=6, key="guid-order"
2017/12/25 18:09:17 redis connected, starting watchloop for "guid-order"
2017/12/25 18:09:17 get step=10001
...
```

```
guid top guid-order                                                                  !10221
10003475506
10003542549
10003350149
10003783992
10003520724
10003975700
10003512200
10003356427
10003851569
10003690851
10003187507
```



Commands:
---------------------

