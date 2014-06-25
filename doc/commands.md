## Total

ledisdb use redis protocol called RESP(REdis Serialization Protocol), [here](http://redis.io/topics/protocol).

ledisdb all commands return RESP fomrat. Later I will use int64 refer RESP integer, string refer RESP simple string, bulk string refer RESP bulk string, and array refer RESP arrays.

## KV 

### decr key
Decrements the number stored at key by one. If the key does not exist, it is set to 0 before decrementing.
An error returns if the value for the key is a wrong type that can not be represented as a signed 64 bit integer.

**Return value**

int64: the value of key after the decrement

**Examples**

```
ledis> decr mykey
(integer) -1
ledis> decr mykey
(integer) -2
ledis> SET mykey "234293482390480948029348230948"
OK
ledis> decr mykey
ERR strconv.ParseInt: parsing "234293482390480948029348230948“: invalid syntax
```

### decrby key decrement

Decrements the number stored at key by decrement. like decr.

**Return value**

int64: the value of key after the decrement

**Examples**

```
ledis> set mykey “10“
OK
ledis> decrby mykey “5“
(integer) 5
```

### del

Removes the specified keys.

**Return value**

int64: The number of input keys 

**Examples**

```
ledis> set key1 "hello"
OK
ledis> set key2 "world"
OK
ledis> del key1 key2
(integer) 2
```

### exists

Returns if key exists

**Return value**

int64, specifically:
- 1 if the key exists.
- 0 if the key does not exists.

**Examples**

```
ledis> set key1 "hello"
OK
ledis> exists key1
(integer) 1
ledis> exists key2
(integer) 0
```

### get

Get the value of key. If the key does not exists, it returns nil value.

**Return value**

bulk: the value of key, or nil when key does not exist.


**Examples**

```

```

### getset
**Return value**

**Examples**

### incr
**Return value**

**Examples**

### incrby
**Return value**

**Examples**
### mget
**Return value**

**Examples**
### mset
**Return value**

**Examples**
### set
**Return value**

**Examples**
### setnx
**Return value**

**Examples**
### expire
**Return value**

**Examples**
### expireat
**Return value**

**Examples**
### ttl
**Return value**

**Examples**
### persist
**Return value**

**Examples**

## Hash

### hdel
**Return value**

**Examples**
### hexists
**Return value**

**Examples**
### hget
**Return value**

**Examples**
### hgetall
**Return value**

**Examples**
### hincrby
**Return value**

**Examples**
### hkeys
**Return value**

**Examples**
### hlen
**Return value**

**Examples**
### hmget
**Return value**

**Examples**
### hmset
**Return value**

**Examples**
### hset
**Return value**

**Examples**
### hvals
**Return value**

**Examples**

## List

### lindex
**Return value**

**Examples**
### llen
**Return value**

**Examples**
### lpop
**Return value**

**Examples**
### lrange
**Return value**

**Examples**
### lpush
**Return value**

**Examples**
### rpop
**Return value**

**Examples**
### rpush
**Return value**

**Examples**
### lclear
**Return value**

**Examples**
### lexpire
**Return value**

**Examples**
### lexpireat
**Return value**

**Examples**
### lttl
**Return value**

**Examples**
### lpersist 
**Return value**

**Examples**

## ZSet

### zadd
**Return value**

**Examples**
### zcard
**Return value**

**Examples**
### zcount
**Return value**

**Examples**
### zincrby
**Return value**

**Examples**
### zrange
**Return value**

**Examples**
### zrangebyscore
**Return value**

**Examples**
### zrank
**Return value**

**Examples**
### zrem
**Return value**

**Examples**
### zremrangebyrank
**Return value**

**Examples**
### zremrangebyscore
**Return value**

**Examples**
### zrevrange
**Return value**

**Examples**
### zrevrangebyscore
**Return value**

**Examples**
### zscore
**Return value**

**Examples**
### zclear
**Return value**

**Examples**
### zexpire
**Return value**

**Examples**
### zexpireat
**Return value**

**Examples**
### zttl
**Return value**

**Examples**
### zpersist
**Return value**

**Examples**

## Replication

### slaveof
**Return value**

**Examples**
### fullsync
**Return value**

**Examples**
### sync
**Return value**

**Examples**

## Server

### ping
**Return value**

**Examples**
### echo
**Return value**

**Examples**
### select
**Return value**

**Examples**
