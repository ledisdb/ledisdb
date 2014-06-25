## Total

ledisdb use redis protocol called RESP(REdis Serialization Protocol), [here](http://redis.io/topics/protocol).

ledisdb all commands return RESP fomrat and it will use int64 instead of  RESP integer, string instead of RESP simple string, bulk string instead of RESP bulk string, and array instead of RESP arrays below.

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

### del key [key ...]

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

### exists key

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

### get key

Get the value of key. If the key does not exists, it returns nil value.

**Return value**

bulk: the value of key, or nil when key does not exist.


**Examples**

```
ledis> get nonexisting
(nil)
ledis> set mykey "hello"
OK
ledis> get mykey
"hello"
```

### getset key value

Atomically sets key to value and returns the old value stored at key.

**Return value**

bulk: the old value stored at key, or nil when key did not exists.

**Examples**

```
ledis> set mykey "hello"
OK
ledis> getset mykey "world"
"hello"
ledis> get mykey
"world"
```

### incr key

Increments the number stored at key by one. If the key does not exists, it is set to 0 before incrementing.

**Return value**

int64: the value of key after the increment

**Examples**

```
ledis> set mykey "10"
OK
ledis> incr mykey
(integer) 11
ledis> get mykey
"11"
```

### incrby key increment

Increments the number stored at key by increment. If the key does not exists, it is set to 0 before incrementing.

**Return value**

int64: the value of key after the increment

**Examples**

```
ledis> set mykey "10"
OK
ledis> incrby mykey 5
(integer) 15
```

### mget key [key ...]

Returns the values of all specified keys. If the key does not exists, a nil will return.

**Return value**

array: list of values at the specified keys

**Examples**

```
ledis> set key1 "hello"
OK
ledis> set key2 "world"
OK
ledis> mget key1 key2 nonexisting
1) "hello"
2) "world"
3) (nil)
```

### mset key value [key value ...]

Sets the given keys to their respective values.

**Return value**

string: always OK

**Examples**

```
ledis> mset key1 "hello" key2 "world"
OK
ledis> get key1
"hello"
ledis> get key2
"world"
```

### set key value

Set key to the value.

**Return value**

string: OK

**Examples**

```
ledis> set mykey "hello"
OK
ledis> get mykey
"hello"
```

### setnx key value

Set key to the value if key does not exist. If key already holds a value, no operation is performed.

**Return value**

int64:

- 1 if the key was set
- 0 if the key was not set

**Examples**

```
ledis> setnx mykey "hello"
(integer) 1
ledis> setnx mykey "world"
(integer) 0
ledis> get mykey
"hello"
```

### expire key seconds

Set a timeout on key. After the timeout has expired, the key will be deleted.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> set mykey "hello"
OK
ledis> expire mykey 60
(integer) 1
ledis> expire mykey 60
(integer) 1
ledis> ttl mykey
(integer) 58
ledis> persist mykey
(integer) 1
```

### expireat key timestamp

Set an expired unix timestamp on key. 

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> SET mykey "Hello"
OK
ledis> EXPIREAT mykey 1293840000
(integer) 1
ledis> EXISTS mykey
(integer) 0
```

### ttl key

Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, -1 returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis> set mykey "hello"
OK
ledis> expire mykey 10
(integer) 1
ledis> ttl mykey
(integer) 8
```

### persist key

Remove the existing timeout on key

**Return value**

int64:

- 1 if the timeout was removed
- 0 if key does not exist or does not have an timeout

**Examples**

```
ledis> set mykey "hello"
OK
ledis> expire mykey 60
(integer) 1
ledis> ttl mykey
(integer) 57
ledis> persist mykey
(integer) 1
ledis> ttl mykey
(integer) -1
```

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
