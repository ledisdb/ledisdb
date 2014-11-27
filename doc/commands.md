## Summary

ledisdb use redis protocol called RESP(REdis Serialization Protocol), [here](http://redis.io/topics/protocol).

ledisdb all commands return RESP format and it will use `int64` instead of  `RESP integer`, `string` instead of `RESP simple string`, `bulk string` instead of `RESP bulk string`, and `array` instead of `RESP arrays` below.

Table of Contents
=================


- [Summary](#summary)
- [KV](#kv)
	- [DECR key](#decr-key)
	- [DECRBY key decrement](#decrby-key-decrement)
	- [DEL key [key ...]](#del-key-key-)
	- [EXISTS key](#exists-key)
	- [GET key](#get-key)
	- [GETSET key value](#getset-key-value)
	- [INCR key](#incr-key)
	- [INCRBY key increment](#incrby-key-increment)
	- [MGET key [key ...]](#mget-key-key-)
	- [MSET key value [key value ...]](#mset-key-value-key-value-)
	- [SET key value](#set-key-value)
	- [SETNX key value](#setnx-key-value)
	- [SETEX key seconds value](#setex-key-seconds-value)
	- [EXPIRE key seconds](#expire-key-seconds)
	- [EXPIREAT key timestamp](#expireat-key-timestamp)
	- [TTL key](#ttl-key)
	- [PERSIST key](#persist-key)
	- [XSCAN key [MATCH match] [COUNT count]](#xscan-key-match-match-count-count)
	- [XREVSCAN key [MATCH match] [COUNT count]](#xrevscan-key-match-match-count-count)
	- [DUMP key](#dump-key)
- [Hash](#hash)
	- [HDEL key field [field ...]](#hdel-key-field-field-)
	- [HEXISTS key field](#hexists-key-field)
	- [HGET key field](#hget-key-field)
	- [HGETALL key](#hgetall-key)
	- [HINCRBY key field increment](#hincrby-key-field-increment)
	- [HKEYS key](#hkeys-key)
	- [HLEN key](#hlen-key)
	- [HMGET key field [field ...]](#hmget-key-field-field-)
	- [HMSET key field value [field value ...]](#hmset-key-field-value-field-value-)
	- [HSET key field value](#hset-key-field-value)
	- [HVALS key](#hvals-key)
	- [HCLEAR key](#hclear-key)
	- [HMCLEAR key [key ...]](#hmclear-key-key)
	- [HEXPIRE key seconds](#hexpire-key-seconds)
	- [HEXPIREAT key timestamp](#hexpireat-key-timestamp)
	- [HTTL key](#httl-key)
	- [HPERSIST key](#hpersist-key)
	- [HXSCAN key [MATCH match] [COUNT count]](#hxscan-key-match-match-count-count)
	- [HXREVSCAN key [MATCH match] [COUNT count]](#hxrevscan-key-match-match-count-count)
	- [HDUMP key](#hdump-key)
- [List](#list)
	- [BLPOP key [key ...] timeout](#blpop-key-key--timeout)
	- [BRPOP key [key ...] timeout](#brpop-key-key--timeout)
	- [LINDEX key index](#lindex-key-index)
	- [LLEN key](#llen-key)
	- [LPOP key](#lpop-key)
	- [LRANGE key start stop](#lrange-key-start-stop)
	- [LPUSH key value [value ...]](#lpush-key-value-value-)
	- [RPOP key](#rpop-keuser-content-y)
	- [RPUSH key value [value ...]](#rpush-key-value-value-)
	- [LCLEAR key](#lclear-key)
	- [LMCLEAR key [key...]](#lmclear-key-key-)
	- [LEXPIRE key seconds](#lexpire-key-seconds)
	- [LEXPIREAT key timestamp](#lexpireat-key-timestamp)
	- [LTTL key](#lttl-key)
	- [LPERSIST key](#lpersist-key)
	- [LXSCAN key [MATCH match] [COUNT count]](#lxscan-key-match-match-count-count)
	- [LXREVSCAN key [MATCH match] [COUNT count]](#lxrevscan-key-match-match-count-count)
	- [LDUMP key](#ldump-key)
- [Set](#set)
	- [SADD key member [member ...]](#sadd-key-member-member-)
	- [SCARD key](#scard-key)
	- [SDIFF key [key ...]](#sdiff-key-key-)
	- [SDIFFSTORE destination key [key ...]](#sdiffstore-destination-key-key-)
	- [SINTER key [key ...]](#sinter-key-key-)
	- [SINTERSTORE destination key [key ...]](#sinterstore-destination-key-key-)
	- [SISMEMBER key member](#sismember-key-member)
	- [SMEMBERS key](#smembers-key)
	- [SREM key member [member ...]](#srem-key-member-member-)
	- [SUNION key [key ...]](#sunion-key-key-)
	- [SUNIONSTORE destination key [key ...]](#sunionstore-destination-key-key-)
	- [SCLEAR key](#sclear-key)
	- [SMCLEAR key [key...]](#smclear-key-key)
	- [SEXPIRE key seconds](#sexpire-key-seconds)
	- [SEXPIREAT key timestamp](#sexpireat-key-timestamp)
	- [STTL key](#sttl-key)
	- [SPERSIST key](#spersist-key)
	- [SXSCAN key [MATCH match] [COUNT count]](#sxscan-key-match-match-count-count)
	- [SXREVSCAN key [MATCH match] [COUNT count]](#sxrevscan-key-match-match-count-count)
	- [SDUMP key](#sdump-key)
- [ZSet](#zset)
	- [ZADD key score member [score member ...]](#zadd-key-score-member-score-member-)
	- [ZCARD key](#zcard-key)
	- [ZCOUNT key min max](#zcount-key-min-max)
	- [ZINCRBY key increment member](#zincrby-key-increment-member)
	- [ZRANGE key start stop [WITHSCORES]](#zrange-key-start-stop-withscores)
	- [ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]](#zrangebyscore-key-min-max-withscores-limit-offset-count)
	- [ZRANK key member](#zrank-key-member)
	- [ZREM key member [member ...]](#zrem-key-member-member-)
	- [ZREMRANGEBYRANK key start stop](#zremrangebyrank-key-start-stop)
	- [ZREMRANGEBYSCORE key min max](#zremrangebyscore-key-min-max)
	- [ZREVRANGE key start stop [WITHSCORES]](#zrevrange-key-start-stop-withscores)
	- [ZREVRANGEBYSCORE  key max min [WITHSCORES] [LIMIT offset count]](#zrevrangebyscore-key-max-min-withscores-limit-offset-count)
	- [ZREVRANK key member](#zrevrank-key-member)
	- [ZSCORE key member](#zscore-key-member)
	- [ZCLEAR key](#zclear-key)
	- [ZMCLEAR key [key ...]](#zmclear-key-key-)
	- [ZEXPIRE key seconds](#zexpire-key-seconds)
	- [ZEXPIREAT key timestamp](#zexpireat-key-timestamp)
	- [ZTTL key](#zttl-key)
	- [ZPERSIST key](#zpersist-key)
    - [ZUNIONSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
](#zunionstore-destination-numkeys-key-key--weights-weight-weight--aggregate-summinmax)
    - [ZINTERSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]
](#zinterstore-destination-numkeys-key-key--weights-weight-weight--aggregate-summinmax)
	- [ZXSCAN key [MATCH match] [COUNT count]](#zxscan-key-match-match-count-count)
	- [ZXREVSCAN key [MATCH match] [COUNT count]](#zxrevscan-key-match-match-count-count)
	- [ZRANGEBYLEX key min max [LIMIT offset count]](#zrangebylex-key-min-max-limit-offset-count)
	- [ZREMRANGEBYLEX key min max](#zremrangebylex-key-min-max)
	- [ZLEXCOUNT key min max](#zlexcount-key-min-max)
	- [ZDUMP key](#zdump-key)
- [Bitmap](#bitmap)
	- [BGET key](#bget-key)
	- [BGETBIT key offset](#bgetbit-key-offset)
	- [BSETBIT key offset value](#bsetbit-key-offset-value)
	- [BMSETBIT key offset value[offset value ...]](#bmsetbit-key-offset-value-offset-value-)
	- [BOPT operation destkey key [key ...]](#bopt-operation-destkey-key-key-)
	- [BCOUNT key [start, end]](#bcount-key-start-end)
	- [BEXPIRE key seconds](#bexpire-key-seconds)
	- [BEXPIREAT key timestamp](#bexpireat-key-timestamp)
	- [BTTL key](#bttl-key)
	- [BPERSIST key](#bpersist-key)
	- [BXSCAN key [MATCH match] [COUNT count]](#bxscan-key-match-match-count-count)
	- [BXREVSCAN key [MATCH match] [COUNT count]](#bxrevscan-key-match-match-count-count)
- [Replication](#replication)
	- [SLAVEOF host port [RESTART] [READONLY]](#slaveof-host-port-restart-readonly)
	- [FULLSYNC [NEW]](#fullsync-new)
	- [SYNC logid](#sync-logid)
- [Server](#server)
	- [PING](#ping)
	- [ECHO message](#echo-message)
	- [SELECT index](#select-index)
	- [FLUSHALL](#flushall)
	- [FLUSHDB](#flushdb)
	- [INFO [section]](#info-section)
	- [TIME](#time)
	- [CONFIG REWRITE](#config-rewrite)
	- [RESTORE key ttl value](#restore-key-ttl-value)
- [Transaction](#transaction)
	- [BEGIN](#begin)
	- [ROLLBACK](#rollback)
	- [COMMIT](#commit)
- [Script](#script)
	- [EVAL script numkeys key [key ...] arg [arg ...]](#eval-script-numkeys-key-key--arg-arg-)
	- [EVALSHA sha1 numkeys key [key ...] arg [arg ...]](#evalsha-sha1-numkeys-key-key--arg-arg-)
	- [SCRIPT LOAD script](#script-load-script)
	- [SCRIPT EXISTS script [script ...]](#script-exists-script-script-)
	- [SCRIPT FLUSH](#script-flush)

## KV 

### DECR key
Decrements the number stored at key by one. If the key does not exist, it is set to 0 before decrementing.
An error returns if the value for the key is a wrong type that can not be represented as a `signed 64 bit integer`.

**Return value**

int64: the value of key after the decrement

**Examples**

```
ledis> DECR mykey
(integer) -1
ledis> DECR mykey
(integer) -2
ledis> SET mykey "234293482390480948029348230948"
OK
ledis> DECR mykey
ERR strconv.ParseInt: parsing "234293482390480948029348230948“: invalid syntax
```


### DECRBY key decrement

Decrements the number stored at key by decrement. like `DECR`.

**Return value**

int64: the value of key after the decrement

**Examples**

```
ledis> SET mykey “10“
OK
ledis> DECRBY mykey “5“
(integer) 5
```

### DEL key [key ...]

Removes the specified keys.

**Return value**

int64: The number of input keys 

**Examples**

```
ledis> SET key1 "hello"
OK
ledis> SET key2 "world"
OK
ledis> DEL key1 key2
(integer) 2
```

### EXISTS key

Returns if key exists

**Return value**

int64, specifically:

- 1 if the key exists.
- 0 if the key does not exists.

**Examples**

```
ledis> SET key1 "hello"
OK
ledis> EXISTS key1
(integer) 1
ledis> EXISTS key2
(integer) 0
```

### GET key

Get the value of key. If the key does not exists, it returns `nil` value.

**Return value**

bulk: the value of key, or nil when key does not exist.


**Examples**

```
ledis> GET nonexisting
(nil)
ledis> SET mykey "hello"
OK
ledis> GET mykey
"hello"
```

### GETSET key value

Atomically sets key to value and returns the old value stored at key.

**Return value**

bulk: the old value stored at key, or nil when key did not exists.

**Examples**

```
ledis> SET mykey "hello"
OK
ledis> GETSET mykey "world"
"hello"
ledis> GET mykey
"world"
```

### INCR key

Increments the number stored at key by one. If the key does not exists, it is SET to `0` before incrementing.

**Return value**

int64: the value of key after the increment

**Examples**

```
ledis> SET mykey "10"
OK
ledis> INCR mykey
(integer) 11
ledis> GET mykey
"11"
```

### INCRBY key increment

Increments the number stored at key by increment. If the key does not exists, it is SET to `0` before incrementing.

**Return value**

int64: the value of key after the increment

**Examples**

```
ledis> SET mykey "10"
OK
ledis> INCRBY mykey 5
(integer) 15
```

### MGET key [key ...]

Returns the values of all specified keys. If the key does not exists, a `nil` will return.

**Return value**

array: list of values at the specified keys

**Examples**

```
ledis> SET key1 "hello"
OK
ledis> SET key2 "world"
OK
ledis> MGET key1 key2 nonexisting
1) "hello"
2) "world"
3) (nil)
```

### MSET key value [key value ...]

Sets the given keys to their respective values.

**Return value**

string: always OK

**Examples**

```
ledis> MSET key1 "hello" key2 "world"
OK
ledis> GET key1
"hello"
ledis> GET key2
"world"
```

### SET key value

Set key to the value.

**Return value**

string: OK

**Examples**

```
ledis> SET mykey "hello"
OK
ledis> GET mykey
"hello"
```

### SETNX key value

Set key to the value if key does not exist. If key already holds a value, no operation is performed.

**Return value**

int64:

- 1 if the key was SET
- 0 if the key was not SET

**Examples**

```
ledis> SETNX mykey "hello"
(integer) 1
ledis> SETNX mykey "world"
(integer) 0
ledis> GET mykey
"hello"
```

### SETEX key seconds value
Set key to hold the string value and set key to timeout after a given number of seconds. This command is equivalent to executing the following commands:

```
SET mykey value
EXPIRE mykey seconds
```

**Return value**

Simple string reply

**Examples**

```
ledis> SETEX mykey 10 "Hello"
OK
ledis> TTL mykey
(integer) 10
ledis> GET mykey
"Hello"
ledis> 
```

### EXPIRE key seconds

Set a timeout on key. After the timeout has expired, the key will be deleted.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> SET mykey "hello"
OK
ledis> EXPIRE mykey 60
(integer) 1
ledis> EXPIRE mykey 60
(integer) 1
ledis> TTL mykey
(integer) 58
ledis> PERSIST mykey
(integer) 1
```

### EXPIREAT key timestamp

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

### TTL key

Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, -1 returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis> SET mykey "hello"
OK
ledis> EXPIRE mykey 10
(integer) 1
ledis> TTL mykey
(integer) 8
```

### PERSIST key

Remove the existing timeout on key

**Return value**

int64:

- 1 if the timeout was removed
- 0 if key does not exist or does not have an timeout

**Examples**

```
ledis> SET mykey "hello"
OK
ledis> EXPIRE mykey 60
(integer) 1
ledis> TTL mykey
(integer) 57
ledis> PERSIST mykey
(integer) 1
ledis> TTL mykey
(integer) -1
```

### XSCAN key [MATCH match] [COUNT count] 

Iterate KV keys incrementally.

Key is the start for the current iteration.
Match is the regexp for checking matched key.
Count is the maximum retrieved elememts number, default is 10.

**Return value**

an array of two values, first value is the key for next iteration, second value is an array of elements.

**Examples**

```
ledis>set a 1
OK
ledis>set b 2
OK
ledis>set c 3
OK
127.0.0.1:6380>xscan "" 
1) ""
2) ["a" "b" "c"]
ledis>xscan "" count 1
1) "a"
2) ["a"]
ledis>xscan "a" count 1
1) "b"
2) ["b"]
ledis>xscan "b" count 1
1) "c"
2) ["c"]
ledis>xscan "c" count 1
1) ""
2) []
```

### XREVSCAN key [MATCH match] [COUNT count] 

Reverse iterate KV keys incrementally.

Key is the start for the current iteration.
Match is the regexp for checking matched key.
Count is the maximum retrieved elememts number, default is 10.

**Return value**

an array of two values, first value is the key for next iteration, second value is an array of elements.

**Examples**

```
ledis>set a 1
OK
ledis>set b 2
OK
ledis>set c 3
OK
127.0.0.1:6380>xrevscan "" 
1) ""
2) ["c" "b" "a"]
ledis>xrevscan "" count 1
1) "c"
2) ["c"]
ledis>xrevscan "c" count 1
1) "b"
2) ["b"]
ledis>xrevscan "b" count 1
1) "a"
2) ["a"]
ledis>xrevscan "a" count 1
1) ""
2) []
```

### DUMP key

Serialize the value stored at key with KV type in a Redis-specific format like RDB and return it to the user. The returned value can be synthesized back into a key using the RESTORE command.

**Return value**

bulk: the serialized value

**Examples**

```
ledis> set mykey 10
OK
ledis>DUMP mykey
"\x00\xc0\n\x06\x00\xf8r?\xc5\xfb\xfb_("
```

## Hash

### HDEL key field [field ...]

Removes the specified fiedls from the hash stored at key.

**Return value**

int64: the number of fields that were removed from the hash.

**Examples**

```
ledis> HSET myhash field1 "foo"
(integer) 1
ledis> HDEL myhash field1 field2
(integer) 1
```

### HEXISTS key field

Returns if field is an existing field in the hash stored at key.

**Return value**

int64:

- 1 if the hash contains field
- 0 if the hash does not contain field, or key does not exist.

**Examples**

```
ledis> HSET myhash field1 "foo"
(integer) 1
ledis> HEXISTS myhash field1 
(integer) 1
ledis> HEXISTS myhash field2
(integer) 0
```

### HGET key field

Returns the value associated with field in the hash stored at key.

**Return value**

bulk: the value associated with field, or `nil`.

**Examples**

```
ledis> HSET myhash field1 "foo"
(integer) 1
ledis> HGET myhash field1
"foo"
ledis> HGET myhash field2
(nil)
```

### HGETALL key

Returns all fields and values of the hash stored at key.

**Return value**

array: list of fields and their values stored in the hash, or an empty list (using nil in ledis-cli)

**Examples**

```
ledis> HSET myhash field1 "hello"
(integer) 1
ledis> HSET myhash field2 "world"
(integer) 1
ledis> HGETALL myhash
1) "field1"
2) "hello"
3) "field2"
4) "world"
```

### HINCRBY key field increment

Increments the number stored at field in the hash stored at key by increment. If key does not exist, a new hash key is created. 
If field does not exists the value is set to 0 before incrementing.

**Return value**

int64: the value at field after the increment.

**Examples**

```
ledis> HINCRBY myhash field 1
(integer) 1
ledis> HGET myhash field
"1"
ledis> HINCRBY myhash field 5
(integer) 6
ledis> HINCRBY myhash field -10
(integer) -4
```

### HKEYS key

Return all fields in the hash stored at key.

**Return value**

array: list of fields in the hash, or an empty list.

**Examples**

```
ledis> HSET myhash field1 "hello"
(integer) 1
ledis> HSET myhash field2 "world"
(integer) 1
ledis> HKEYS myhash
1) "field1"
2) "field2"
```

### HLEN key

Returns the number of fields contained in the hash stored at key

**Return value**

int64: number of fields in the hash, or 0 when key does not exist.

**Examples**

```
ledis> HSET myhash field1 "hello"
(integer) 1
ledis> HSET myhash field2 "world"
(integer) 1
ledis> HLEN myhash
(integer) 2
```

### HMGET key field [field ...]

Returns the values associated with the specified fields in the hash stored at key. If field does not exist in the hash, a `nil` value is returned.

**Return value**

array: list of values associated with the given fields.

**Examples**

```
ledis> HSET myhash field1 "hello"
(integer) 1
ledis> HSET myhash field2 "world"
(integer) 1
ledis> HMGET myhash field1 field2 nofield
1) "hello"
2) "world"
3) (nil)
```

### HMSET key field value [field value ...]

Sets the specified fields to their respective values in the hash stored at key.

**Return value**

string: OK

**Examples**

```
ledis> HMSET myhash field1 "hello" field2 "world"
OK
ledis> HMGET myhash field1 field2
1) "hello"
2) "world"
```

### HSET key field value

Sets field in the hash stored at key to value. If key does not exists, a new hash key is created.

**Return value**

int64:

- 1 if field is a new field in the hash and value was set.
- 0 if field already exists in the hash and the value was updated.

**Examples**

```
ledis> HSET myhash field1 "hello"
(integer) 1
ledis> HGET myhash field1
"hello"
ledis> HSET myhash field1 "world"
(integer) 0
ledis> HGET myhash field1
"world"
```

### HVALS key

Returns all values in the hash stored at key.

**Return value**

array: list of values in the hash, or an empty list.

**Examples**

```
ledis> HSET myhash field1 "hello"
(integer) 1
ledis> HSET myhash field2 "world"
(integer) 1
ledis> HVALS myhash
1) "hello"
2) "world"
```

### HCLEAR key 

Deletes the specified hash key

**Return value**

int64: the number of fields in the hash stored at key

**Examples**

```
ledis> HMSET myhash field1 "hello" field2 "world"
OK
ledis> HCLEAR myhash
(integer) 2
```

### HMCLEAR key [key...]

Deletes the specified hash keys.

**Return value**

int64: the number of input keys

**Examples**

```
ledis> HMSET myhash field1 "hello" field2 "world"
OK
ledis> HMCLEAR myhash
(integer) 1
```

### HEXPIRE key seconds

Sets a hash key's time to live in seconds, like expire similarly.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set


**Examples**

```
ledis> HSET myhash a  100
(integer) 1
ledis> HGET myhash a
100
ledis> HEXPIRE myhash 100
(integer) 1
ledis> HTTL myhash
(integer) 94
ledis> HPERSIST myhash
(integer) 1
ledis> HTTL myhash
(integer) -1
ledis> HEXPIRE not_exists_key 100
(integer) 0
```

### HEXPIREAT key timestamp

Sets the expiration for a hash key as a unix timestamp, like expireat similarly.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> HSET myhash a  100
(integer) 1
ledis> HEXPIREAT myhash 1404999999
(integer) 1
ledis> HTTL myhash
(integer) 802475
ledis> HEXPIREAT not_exists_key  1404999999
(integer) 0
```

### HTTL key

Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, `-1` returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis> HSET myhash a  100
(integer) 1
ledis> HEXPIREAT myhash 1404999999
(integer) 1
ledis> HTTL myhash
(integer) 802475
ledis> HTTL not_set_timeout
(integer) -1
```

### HPERSIST key

Remove the expiration from a hash key, like persist similarly.
Remove the existing timeout on key.

**Return value**

int64:

- 1 if the timeout was removed
- 0 if key does not exist or does not have an timeout

```
ledis> HSET myhash a  100
(integer) 1
ledis> HEXPIREAT myhash 1404999999
(integer) 1
ledis> HTTL myhash
(integer) 802475
ledis> HPERSIST myhash
(integer) 1
ledis> HTTL myhash
(integer) -1
ledis> HPERSIST not_exists_key
(integer) 0
```

### HXSCAN key [MATCH match] [COUNT count] 

Iterate Hash keys incrementally.

See [XSCAN](#xscan-key-match-match-count-count) for more information.

### HXREVSCAN key [MATCH match] [COUNT count] 

Reverse iterate Hash keys incrementally.

See [XREVSCAN](#xrevscan-key-match-match-count-count) for more information.

### HDUMP key

See [DUMP](#dump-key) for more information.

## List

### BLPOP key [key ...] timeout

BLPOP is a blocking list pop primitive. It is the blocking version of LPOP because it blocks the connection when there are no elements to pop from any of the given lists. An element is popped from the head of the first list that is non-empty, with the given keys being checked in the order that they are given.

When BLPOP causes a client to block and a non-zero timeout is specified, the client will unblock returning a nil multi-bulk value when the specified timeout has expired.
The timeout argument is interpreted as an double value specifying the maximum number of seconds to block. You can use 0.005 format to support milliseconds timeout.

A timeout of zero can be used to block indefinitely.

BLPOP and BRPOP can not work correctly in transaction now!

**Return value**

array: 

+ A nil multi-bulk when no element could be popped and the timeout expired.
+ A two-element multi-bulk with the first element being the name of the key where an element was popped and the second element being the value of the popped element.

**Examples**
```
ledis> RPUSH list1 a b c
(integer) 3
ledis> BLPOP list1 list2 0
1) "list1"
2) "a"
```

### BRPOP key [key ...] timeout

See [BLPOP key [key ...] timeout](#blpop-key-key--timeout) for more information.

### LINDEX key index
Returns the element at index index in the list stored at key. The index is zero-based, so 0 means the first element, 1 the second element and so on. Negative indices can be used to designate elements starting at the tail of the list. Here, `-1` means the last element, `-2` means the penultimate and so forth.
When the value at key is not a list, an error is returned.

**Return value**

string: the requested element, or `nil` when index is out of range.

**Examples**

```
ledis> RPUSH a 1 2 3
(integer) 3
ledis> LINDEX a 0
1
ledis> LINDEX a 1
2
ledis> LINDEX a 2
3
ledis> LINDEX a 3
(nil)
ledis> LINDEX a -1
3
```

### LLEN key
Returns the length of the list stored at key. If key does not exist, it is interpreted as an empty list and `0`is returned. An error is returned when the value stored at key is not a list.

**Return value**

int64: the length of the list at key.

**Examples**

```
ledis> RPUSH a 'foo'
(integer) 1
ledis> RPUSH a 'bar'
(integer) 2
ledis> LLEN a
(integer) 2
```

### LPOP key
Removes and returns the first element of the list stored at key.

**Return value**

bulk: the value of the first element, or `nil` when key does not exist.

**Examples**

```
ledis> RPUSH a 'one'
(integer) 1
ledis> RPUSH a 'two'
(integer) 2
ledis> RPUSH a 'three'
(integer) 3
ledis> LPOP a
one
```

### LRANGE key start stop
Returns the specified elements of the list stored at key. The offsets start and stop are zero-based indexes, with 0 being the first element of the list (the head of the list), `1` being the next element and so on.

**Return value**

array: list of elements in the specified range.

**Examples**

```
ledis> RPUSH a 'one' 'two' 'three'
(integer) 3
ledis> LRANGE a 0 0
1) "one"
ledis> LRANGE a -100 100
1) "one"
2) "two"
3) "three"
ledis> LRANGE a -3 2
1) "one"
2) "two"
3) "three"
ledis> LRANGE a 0 -1
(empty list or set)
```

### LPUSH key value [value ...]
Insert all the specified values at the head of the list stored at key. If key does not exist, it is created as empty list before performing the push operations. When key holds a value that is not a list, an error is returned.

**Return value**

int64: the length of the list after the push operations.

**Examples**

```
ledis> LPUSH a 1
(integer) 1
ledis> LPUSH a 2
(integer) 2
ledis> LRANGE a 0 2
1) "2"
2) "1"
```

### RPOP key
Removes and returns the last element of the list stored at key.

**Return value**

bulk: the value of the last element, or `nil` when key does not exist.

**Examples**

```
edis > RPUSH a 1
(integer) 1
ledis> RPUSH a 2
(integer) 2
ledis> RPUSH a 3
(integer) 3
ledis> RPOP a
3
ledis> LRANGE a 0 3
1) "1"
2) "2"
```

### RPUSH key value [value ...]
Insert all the specified values at the tail of the list stored at key. If key does not exist, it is created as empty list before performing the push operation. When key holds a value that is not a list, an error is returned.

**Return value**

int64: the length of the list after the push operation.

**Examples**

```
ledis>  RPUSH a 'hello'
(integer) 1
ledis> RPUSH a 'world'
(integer) 2
ledis> LRANGE a 0 2
1) "hello"
2) "world"
```

### LCLEAR key
Deletes the specified list key

**Return value**

int64: the number of values in the list stored at key

**Examples**

```
ledis> RPUSH a 1 2 3
(integer) 3
ledis> LLEN a
(integer) 3
ledis> LCLEAR a
(integer) 3
ledis> LLEN a
(integer) 0
```

### LMCLEAR key [key ...]
Delete multiple keys from list


**Return value**

int64: the number of input keys

**Examples**

```
ledis> rpush a 1
(integer) 1
ledis> rpush b 2
(integer) 1
ledis> lmclear a b
(integer) 2
```

### LEXPIRE key seconds
Set a timeout on key. After the timeout has expired, the key will be deleted.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> RPUSH a 1
(integer) 1
ledis> LEXPIRE a 100
(integer) 1
ledis> LTTL a
(integer) 96
ledis> LPERSIST a
(integer) 1
ledis> LTTL a
(integer) -1
```

### LEXPIREAT key timestamp
Set an expired unix timestamp on key. 

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> RPUSH a 1
(integer) 1
ledis> LEXPIREAT a 1404140183
(integer) 1
ledis> LTTL a
(integer) 570
ledis> LPERSIST a
(integer) 1
ledis> LTTL a
(integer) -1
ledis>
```

### LTTL key
Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, `-1` returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis> RPUSH a 1
(integer) 1
ledis> LEXPIREAT a 1404140183
(integer) 1
ledis> LTTL a
(integer) 570
ledis> LPERSIST a
(integer) 1
ledis> LTTL a
(integer) -1
```

### LPERSIST key
Remove the existing timeout on key

**Return value**

int64:

- 1 if the timeout was removed
- 0 if key does not exist or does not have an timeout

**Examples**

```
ledis> RPUSH a 1
(integer) 1
ledis> LEXPIREAT a 1404140183
(integer) 1
ledis> LTTL a
(integer) 570
ledis> LPERSIST a
(integer) 1
ledis> LTTL a
(integer) -1
ledis> LPERSIST b
(integer) 0
```

### LXSCAN key [MATCH match] [COUNT count] 

Iterate list keys incrementally.

See [XSCAN](#xscan-key-match-match-count-count) for more information.

### LXREVSCAN key [MATCH match] [COUNT count] 

Reverse iterate list keys incrementally.

See [XREVSCAN](#xrevscan-key-match-match-count-count) for more information.

### LDUMP key

See [DUMP](#dump-key) for more information.


## Set

### SADD key member [member ...]
Add the specified members to the set stored at key. Specified members that are already a member of this set are ignored. If key does not exist, a new set is created before adding the specified members.

**Return value**

int64: the number of elements that were added to the set, not including all the elements already present into the set.

**Examples**

```
ledis> SADD myset hello
(integer) 1
ledis> SADD myset world
(integer) 1
ledis> SADD myset hello
(integer) 0
ledis> SMEMBERS myset
1) "hello"
2) "world"
```


### SCARD key

Returns the set cardinality (number of elements) of the set stored at key.

**Return value**

int64: the cardinality (number of elements) of the set, or 0 if key does not exist.

**Examples**

```
ledis> SADD myset hello
(integer) 1
ledis> SADD myset world
(integer) 1
ledis> SADD myset hello
(integer) 0
ledis> SCARD myset
(integer) 2
```


### SDIFF key [key ...]
Returns the members of the set resulting from the difference between the first set and all the successive sets.
For example:

```
key1 = {a,b,c,d}
key2 = {c}
key3 = {a,c,e}
SDIFF key1 key2 key3 = {b,d}
```

Keys that do not exist are considered to be empty sets.


**Return value**

bulk: list with members of the resulting set.

**Examples**

```
ledis> SADD key1 a b c 
(integer) 3
ledis> SADD key2 c d e
(integer) 3
ledis> SDIFF key1 key2
1) "a"
2) "b"
ledis> SDIFF key2 key1
1) "d"
2) "e"
```

### SDIFFSTORE destination key [key ...]
This command is equal to `SDIFF`, but instead of returning the resulting set, it is stored in destination.
If destination already exists, it is overwritten.

**Return value**

int64:  the number of elements in the resulting set.

**Examples**

```
ledis> SADD key1 a b c 
(integer) 3
ledis> SADD key2 c d e
(integer) 3
ledis> SDIFF key1 key2
1) "a"
2) "b"
ledis> SDIFFSTORE key key1 key2
(integer) 2
ledis> SMEMBERS key
1) "a"
2) "b"
```

### SINTER key [key ...]

Returns the members of the set resulting from the intersection of all the given sets.
For example:

```
key1 = {a,b,c,d}
key2 = {c}
key3 = {a,c,e}
SINTER key1 key2 key3 = {c}
```

Keys that do not exist are considered to be empty sets. With one of the keys being an empty set, the resulting set is also empty (since set intersection with an empty set always results in an empty set).

**Return value**

bulk: list with members of the resulting set.

**Examples**

```
ledis> SADD key1 a b c 
(integer) 3
ledis> SADD key2 c d e
(integer) 3
ledis> SINTER key1 key2
1) "c"
ledis> SINTER key2 key_empty
(nil)
```


### SINTERSTORE  destination key [key ...]

This command is equal to `SINTER`, but instead of returning the resulting set, it is stored in destination.
If destination already exists, it is overwritten.

**Return value**

int64: the number of elements in the resulting set.

**Examples**

```
ledis> SADD key1 a b c 
(integer) 3
ledis> SADD key2 c d e
(integer) 3
ledis> SINTERSTORE key key1 key2
(integer) 1
ledis> SMEMBERS key
1) "c"
```


### SISMEMBER  key member
Returns if member is a member of the set stored at key.

**Return value**

Int64 reply, specifically:

- 1 if the element is a member of the set.
- 0 if the element is not a member of the set, or if key does not exist.

**Examples**

```
ledis> SADD myset hello
(integer) 1
ledis> SISMEMBER myset hello
(integer) 1
ledis> SISMEMBER myset hell
(integer) 0
```

### SMEMBERS key 
Returns all the members of the set value stored at key.
This has the same effect as running `SINTER` with one argument key.

**Return value**

bulk: all elements of the set.

**Examples**

```
ledis> SADD myset hello
(integer) 1
ledis> SADD myset world
(integer) 1
ledis> SMEMBERS myset
1) "hello"
2) "world"
```

### SREM  key member [member ...]

Remove the specified members from the set stored at key. Specified members that are not a member of this set are ignored. If key does not exist, it is treated as an empty set and this command returns 0.

**Return value**

int64: the number of members that were removed from the set, not including non existing members.

**Examples**

```
ledis> SADD myset one
(integer) 1
ledis> SADD myset two
(integer) 1
ledis> SADD myset three
(integer) 1
ledis> SREM myset one
(integer) 1
ledis> SREM myset four
(integer) 0
ledis> SMEMBERS myset
1) "three"
2) "two"
```

### SUNION key [key ...]

Returns the members of the set resulting from the union of all the given sets.
For example:

```
key1 = {a,b,c,d}
key2 = {c}
key3 = {a,c,e}
SUNION key1 key2 key3 = {a,b,c,d,e}
```
Keys that do not exist are considered to be empty sets.


**Return value**

bulk: list with members of the resulting set.

**Examples**

```
ledis> SMEMBERS key1
1) "a"
2) "b"
3) "c"
ledis> SMEMBERS key2
1) "c"
2) "d"
3) "e"
ledis> SUNION key1 key2
1) "a"
2) "b"
3) "c"
4) "d"
5) "e"
```

### SUNIONSTORE destination key [key]

This command is equal to SUNION, but instead of returning the resulting set, it is stored in destination.
If destination already exists, it is overwritten.

**Return value**

int64: the number of elements in the resulting set.

**Examples**

```
ledis> SMEMBERS key1
1) "a"
2) "b"
3) "c"
ledis> SMEMBERS key2
1) "c"
2) "d"
3) "e"
ledis> SUNIONSTORE key key1 key2
(integer) 5
ledis> SMEMBERS key
1) "a"
2) "b"
3) "c"
4) "d"
5) "e"
```


### SCLEAR key

Deletes the specified set key

**Return value**

int64: the number of fields in the hash stored at key

**Examples**

```
ledis> SMEMBERS key
1) "a"
2) "b"
3) "c"
4) "d"
5) "e"
ledis> SCLEAR key
(integer) 5
```

### SMCLEAR key [key ...]

Deletes the specified set keys.

**Return value**

int64: the number of input keys

**Examples**

```
ledis> SMCLEAR key1 key2
(integer) 2
ledis> SMCLEAR em1 em2
(integer) 2
```

### SEXPIRE key seconds

Sets a set key’s time to live in seconds, like expire similarly.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> SADD key 1 2 
(integer) 2
ledis> SEXPIRE key 100
(integer) 1
ledis> STTL key
(integer) 95
```


### SEXPIREAT key timestamp

Sets the expiration for a set key as a unix timestamp, like expireat similarly.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> SADD key 1 2 
(integer) 2
ledis> SEXPIREAT key 1408094999
(integer) 1
ledis> STTL key
(integer) 908
```


### STTL key

Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, -1 returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis> SADD key 1 2 
(integer) 2
ledis> SEXPIREAT key 1408094999
(integer) 1
ledis> STTL key
(integer) 908
```


### SPERSIST key 
Remove the expiration from a set key, like persist similarly. Remove the existing timeout on key.

**Return value**

int64:

- 1 if the timeout was removed
- 0 if key does not exist or does not have an timeout

**Examples**

```
ledis> SEXPIREAT key 1408094999
(integer) 1
ledis> STTL key
(integer) 908
ledis> SPERSIST key
(integer) 1
ledis> STTL key
(integer) -1
```

### SXSCAN key [MATCH match] [COUNT count] 

Iterate Set keys incrementally.

See [XSCAN](#xscan-key-match-match-count-count) for more information.


### SXREVSCAN key [MATCH match] [COUNT count] 

Reverse iterate Set keys incrementally.

See [XREVSCAN](#xrevscan-key-match-match-count-count) for more information.

### SDUMP key

See [DUMP](#dump-key) for more information.

## ZSet

### ZADD key score member [score member ...]
Adds all the specified members with the specified scores to the sorted set stored at key. It is possible to specify multiple `score / member` pairs. If a specified member is already a member of the sorted set, the score is updated and the element reinserted at the right position to ensure the correct ordering.

If key does not exist, a new sorted set with the specified members as sole members is created, like if the sorted set was empty. If the key exists but does not hold a sorted set, an error is returned.

The score values should be the string representation of an `int64` number. `+inf` and `-inf` values are valid values as well.

**Currently, we only support int64 type, not double type.**

**Return value**

int64, specifically:

The number of elements added to the sorted sets, **not** including elements already existing for which the score was updated.


**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZADD myzset 1 'uno'
(integer) 1
ledis> ZADD myzset 2 'two' 3 'three'
(integer) 2
ledis> ZRANGE myzset 0 -1 WITHSCORES
1) "one"
2) "1"
3) "uno"
4) "1"
5) "two"
6) "2"
7) "three"
8) "3"
```

### ZCARD key
Returns the sorted set cardinality (number of elements) of the sorted set stored at key.

**Return value**

int64: the cardinality (number of elements) of the sorted set, or `0` if key does not exist.

**Examples**

```
edis > ZADD myzset 1 'one'
(integer) 1
ledis> ZADD myzset 1 'uno'
(integer) 1
ledis> ZADD myzset 2 'two' 3 'three'
(integer) 2
ledis> ZRANGE myzset 0 -1 WITHSCORES
1) "one"
2) "1"
3) "uno"
4) "1"
5) "two"
6) "2"
7) "three"
8) "3"
ledis> ZCARD myzset
(integer) 4
```

### ZCOUNT key min max
Returns the number of elements in the sorted set at key with a score between `min` and `max`.
The `min` and `max` arguments have the same semantic as described for `ZRANGEBYSCORE`.

**Return value**

int64: the number of elements in the specified score range.

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZADD myzset 1 'uno'
(integer) 1
ledis> ZADD myzset 2 'two' 3 'three'
(integer) 2
ledis> ZRANGE myzset 0 -1 WITHSCORES
1) "one"
2) "1"
3) "uno"
4) "1"
5) "two"
6) "2"
7) "three"
8) "3"
ledis> ZCOUNT myzset -inf +inf
(integer) 4
ledis> ZCOUNT myzset (1 3
(integer) 2
```

### ZINCRBY key increment member

Increments the score of member in the sorted set stored at key by increment. If member does not exist in the sorted set, it is added with increment as its score (as if its previous score was 0). If key does not exist, a new sorted set with the specified member as its sole member is created.
An error is returned when key exists but does not hold a sorted set.
The score value should be the string representation of a numeric value. It is possible to provide a negative value to decrement the score.

**Return value**

bulk: the new score of member (an int64 number), represented as string.

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZADD myzset 2 'two'
(integer) 1
ledis> ZINCRBY myzset 2 'one'
3
ledis> ZRANGE myzset 0 -1 WITHSCORES
1) "two"
2) "2"
3) "one"
4) "3"
```

### ZRANGE key start stop [WITHSCORES]
Returns the specified range of elements in the sorted set stored at key. The elements are considered to be ordered from the lowest to the highest score. Lexicographical order is used for elements with equal score.

**Return value**

array: list of elements in the specified range (optionally with their scores).

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZADD myzset 2 'two'
(integer) 1
ledis> ZADD myzset 3 'three'
(integer) 1
ledis> ZRANGE myzset 0 -1
1) "one"
2) "two"
3) "three"
ledis> ZRANGE myzset 2 3
1) "three"
ledis> ZRANGE myzset -2 -1
1) "two"
2) "three"
```

### ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]

Returns all the elements in the sorted set at key with a score between `min` and `max` (including elements with score equal to `min` or `max`). The elements are considered to be ordered from low to high scores.

**Exclusive intervals and infinity**

`min` and `max` can be `-inf` and `+inf`, so that you are not required to know the highest or lowest score in the sorted set to get all elements from or up to a certain score.
By default, the interval specified by min and max is closed (inclusive). It is possible to specify an open interval (exclusive) by prefixing the score with the character (. For example:

```
ZRANGEBYSCORE zset (1 5
```

Will return all elements with 1 < score <= 5 while:

```
ZRANGEBYSCORE zset (5 (10
```

Will return all the elements with 5 < score < 10 (5 and 10 excluded).


**Return value**

array: list of elements in the specified score range (optionally with their scores).

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZADD myzset 2 'two'
(integer) 1
ledis> ZADD myzset 3 'three'
(integer) 1
ledis> ZRANGEBYSCORE myzset -inf +inf WITHSCORES
1) "one"
2) "1"
3) "two"
4) "2"
5) "three"
6) "3"
ledis> ZRANGEBYSCORE myzset -inf +inf WITHSCORES LIMIT  2 5
1) "three"
2) "3"
ledis> ZRANGEBYSCORE myzset (1 2 WITHSCORES
1) "two"
2) "2"
ledis> ZRANGEBYSCORE myzset (1 (2 WITHSCORES
```

### ZRANK key member
Returns the rank of member in the sorted set stored at key, with the scores ordered from low to high. The rank (or index) is `0-based`, which means that the member with the lowest score has rank 0.

**Return value**

Return value

- If member exists in the sorted set, Integer reply: the rank of member.
- If member does not exist in the sorted set or key does not exist, Bulk string reply: nil.

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZADD myzset 2 'two'
(integer) 1
ledis> ZADD myzset 3 'three'
(integer) 1
ledis> ZRANGEBYSCORE  myzset -inf +inf WITHSCORES
1) "one"
2) "1"
3) "two"
4) "2"
5) "three"
6) "3"
ledis> ZRANK myzset 'three'
(integer) 2
```


### ZREM key member [member ...]
Removes the specified members from the sorted set stored at key. Non existing members are ignored.
An error is returned when key exists and does not hold a sorted set.

**Return value**

int64 reply, specifically:

The number of members removed from the sorted set, not including non existing members.

**Examples**

```
ledis> ZADD myzset 1 one 2 two 3 three 4 four
(integer) 3
ledis> ZRANGE myzset 0 -1
1) "one"
2) "two"
3) "three"
4) "four"
ledis> ZREM myzset three
(integer) 1
ledis> ZREM myzset one four three
(integer) 2
```

### ZREMRANGEBYRANK key start stop
Removes all elements in the sorted set stored at key with rank between start and stop. Both start and stop are 0 -based indexes with 0 being the element with the lowest score. These indexes can be negative numbers, where they indicate offsets starting at the element with the highest score. For example: -1 is the element with the highest score, -2 the element with the second highest score and so forth.

**Return value**

int64: the number of elements removed.

**Examples**

```
ledis> ZADD myzset 1 one 2 two 3 three 4 four
(integer) 3
ledis> ZREMRANGEBYRANK myzset 0 2
(integer) 3
ledis> ZRANGE myzset 0 -1 WITHSCORES
1) "four"
2) "4"
```


### ZREMRANGEBYSCORE key min max
Removes all elements in the sorted set stored at key with a score between `min` and `max` (inclusive). `Min` and `max` can be exclusive, following the syntax of `ZRANGEBYSCORE`.

**Return value**

int64: the number of elements removed.

**Examples**

```
ledis> ZADD myzset 1 one 2 two 3 three 4 four
(integer) 4
ledis> ZREMRANGEBYSCORE myzset -inf (2
(integer) 1
ledis> ZRANGE myzset 0 -1 WITHSCORES
1) "two"
2) "2"
3) "three"
4) "3"
5) "four"
6) "4"
```

### ZREVRANGE key start stop [WITHSCORES]
Returns the specified range of elements in the sorted set stored at key. The elements are considered to be ordered from the highest to the lowest score. Descending lexicographical order is used for elements with equal score.
Apart from the reversed ordering, ZREVRANGE is similar to `ZRANGE`.

**Return value**

array: list of elements in the specified range (optionally with their scores).

**Examples**

```
ledis> ZADD myzset 1 one 2 two 3 three 4 four
(integer) 4
ledis> ZREVRANGE myzset 0 -1
1) "four"
2) "three"
3) "two"
4) "one"
```

### ZREVRANGEBYSCORE  key max min [WITHSCORES] [LIMIT offset count]
Returns all the elements in the sorted set at key with a score between max and min (including elements with score equal to max or min). In contrary to the default ordering of sorted sets, for this command the elements are considered to be ordered from high to low scores.
The elements having the same score are returned in reverse lexicographical order.
Apart from the reversed ordering, ZREVRANGEBYSCORE is similar to ZRANGEBYSCORE.

**Return value**

array: list of elements in the specified score range (optionally with their scores).

**Examples**

```
ledis>  ZADD myzset 1 one 2 two 3 three 4 four
(integer) 4
ledis> ZREVRANGEBYSCORE myzset +inf -inf
1) "four"
2) "three"
3) "two"
4) "one"
ledis> ZREVRANGEBYSCORE myzset 2 1
1) "two"
2) "one"
ledis> ZREVRANGEBYSCORE myzset 2 (1
1) "two"
ledis> ZREVRANGEBYSCORE myzset (2 (1
(empty list or set)
ledis> ZREVRANGEBYSCORE myzset +inf -inf WITHSCORES LIMIT 1 2
1) "three"
2) "3"
3) "two"
4) "2"
```

### ZREVRANK key member
Returns the rank of member in the sorted set stored at key, with the scores ordered from high to low. The rank (or index) is 0-based, which means that the member with the highest score has rank 0.
Use ZRANK to get the rank of an element with the scores ordered from low to high.

**Return value**

- If member exists in the sorted set, Integer reply: the rank of member.
- If member does not exist in the sorted set or key does not exist, Bulk string reply: nil.


**Examples**

```
ledis> ZADD myzset 1 one
(integer) 1
ledis> ZADD myzset 2 two
(integer) 1
ledis> ZREVRANK myzset one
(integer) 1
ledis> ZREVRANK myzset three
(nil)
```


### ZSCORE key member
Returns the score of member in the sorted set at key.
If member does not exist in the sorted set, or key does not exist, `nil` is returned.

**Return value**

bulk: the score of member (an `int64` number), represented as string.

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZSCORE myzset 'one'
1
```

### ZCLEAR key
Delete the specified  key

**Return value**

int64: the number of members in the zset stored at key

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZADD myzset 2 'two'
(integer) 1
ledis> ZADD myzset 3 'three'
(integer) 1
ledis> ZRANGE myzset 0 -1
1) "one"
2) "two"
3) "three"
ledis> ZCLEAR myzset
(integer) 3
```

### ZMCLEAR key [key ...]
Delte multiple keys one time.

**Return value**

int64: the number of input keys

**Examples**

```
ledis> ZADD myzset1 1 'one'
(integer) 1
ledis> ZADD myzset2 2 'two'
(integer) 1
ledis> ZMCLEAR myzset1 myzset2
(integer) 2
```

### ZEXPIRE key seconds

Set a timeout on key. After the timeout has expired, the key will be deleted.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set


**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZEXPIRE myzset 100
(integer) 1
ledis> ZTTL myzset
(integer) 97
ledis> ZPERSIST myzset
(integer) 1
ledis> ZTTL mset
(integer) -1
ledis> ZEXPIRE myzset1 100
(integer) 0
```

### ZEXPIREAT key timestamp
Set an expired unix timestamp on key. Similar to ZEXPIRE.

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZEXPIREAT myzset 1404149999
(integer) 1
ledis> ZTTL myzset
(integer) 7155
ledis> ZPERSIST myzset
(integer) 1
ledis> ZTTL mset
(integer) -1
ledis> ZEXPIREAT myzset1 1404149999
(integer) 0
```


### ZTTL key
Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, `-1` returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZEXPIRE myzset 100
(integer) 1
ledis> ZTTL myzset
(integer) 97
ledis> ZTTL myzset2
(integer) -1
```

### ZPERSIST key
Remove the existing timeout on key.

**Return value**

int64:

- 1 if the timeout was removed
- 0 if key does not exist or does not have an timeout

**Examples**

```
ledis> ZADD myzset 1 'one'
(integer) 1
ledis> ZEXPIRE myzset 100
(integer) 1
ledis> ZTTL myzset
(integer) 97
ledis> ZPERSIST myzset
(integer) 1
ledis> ZTTL mset
(integer) -1
```

### ZUNIONSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]

Computes the union of numkeys sorted sets given by the specified keys, and stores the result in destination. It is mandatory to provide the number of input keys (numkeys) before passing the input keys and the other (optional) arguments.

By default, the resulting score of an element is the sum of its scores in the sorted sets where it exists.
Using the WEIGHTS option, it is possible to specify a multiplication factor for each input sorted set. This means that the score of every element in every input sorted set is multiplied by this factor before being passed to the aggregation function. When WEIGHTS is not given, the multiplication factors default to 1.

With the AGGREGATE option, it is possible to specify how the results of the union are aggregated. This option defaults to SUM, where the score of an element is summed across the inputs where it exists. When this option is set to either MIN or MAX, the resulting set will contain the minimum or maximum score of an element across the inputs where it exists.

If destination already exists, it is overwritten.


**Return value**

int64: the number of elements in the resulting sorted set at destination.

**Examples**

```
ledis> ZADD zset1 1 "one"
(interger) 1
ledis> ZADD zset1 2 "two"
(interger) 1
ledis> ZADD zset2 1 "one"
(interger) 1
ledis> ZADD zset2 2 "two"
(interger) 1
ledis> ZADD zset2 3 "three"
(interger) 1
ledis> ZUNIONSTORE out 2 zset1 zset2 WEIGHTS 2 3
(interger) 3
ledis> ZRANGE out 0 -1 WITHSCORES
1) "one"
2) "5"
3) "three"
4) "9"
5) "two"
6) "10"
```

### ZINTERSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX]

Computes the intersection of numkeys sorted sets given by the specified keys, and stores the result in destination. It is mandatory to provide the number of input keys (numkeys) before passing the input keys and the other (optional) arguments.

By default, the resulting score of an element is the sum of its scores in the sorted sets where it exists. Because intersection requires an element to be a member of every given sorted set, this results in the score of every element in the resulting sorted set to be equal to the number of input sorted sets.

For a description of the `WEIGHTS` and `AGGREGATE` options, see [ZUNIONSTORE](#zunionstore-destination-numkeys-key-key--weights-weight-weight--aggregate-summinmax).

If destination already exists, it is overwritten.



**Return value**

int64: the number of elements in the resulting sorted set at destination.

**Examples**

```
ledis> ZADD zset1 1 "one"
(interger) 1
ledis> ZADD zset1 2 "two"
(interger) 1
ledis> ZADD zset2 1 "one"
(interger) 1
ledis> ZADD zset2 2 "two"
(interger) 1
ledis> ZADD zset2 3 "three"
(interger) 1
ledis> ZINTERSTORE out 2 zset1 zset2 WEIGHTS 2 3
(interger) 3
ledis> ZRANGE out 0 -1 WITHSCORES
1) "one"
2) "5"
3) "two"
4) "10"
```

### ZXSCAN key [MATCH match] [COUNT count] 

Iterate ZSet keys incrementally.

See [XSCAN](#xscan-key-match-match-count-count) for more information.

### ZXREVSCAN key [MATCH match] [COUNT count] 

Reverse iterate ZSet keys incrementally.

See [XREVSCAN](#xrevscan-key-match-match-count-count) for more information.

### ZRANGEBYLEX key min max [LIMIT offset count]

When all the elements in a sorted set are inserted with the same score, in order to force lexicographical ordering, this command returns all the elements in the sorted set at key with a value between min and max.

If the elements in the sorted set have different scores, the returned elements are unspecified.

Valid start and stop must start with ( or [, in order to specify if the range item is respectively exclusive or inclusive. The special values of + or - for start and stop have the special meaning or positively infinite and negatively infinite strings, so for instance the command ZRANGEBYLEX myzset - + is guaranteed to return all the elements in the sorted set, if all the elements have the same score.

**Return value**

array: list of elements in the specified score range

**Example**

```
ledis> ZADD myzset 0 a 0 b 0 c 0 d 0 e 0 f 0 g
(integer) 7
ledis> ZRANGEBYLEX myzset - [c
1) "a"
2) "b"
3) "c"
ledis> ZRANGEBYLEX myzset - (c
1) "a"
2) "b"
ledis> ZRANGEBYLEX myzset [aaa (g
1) "b"
2) "c"
3) "d"
4) "e"
5) "f"
```

### ZREMRANGEBYLEX key min max

Removes all elements in the sorted set stored at key between the lexicographical range specified by min and max.

**Return value**

int64: he number of elements removed.

**Example**

```
ledis> ZADD myzset 0 a 0 b 0 c 0 d 0 e 0 f 0 g
(integer) 7
ledis> ZREMRANGEBYLEX myzset - [c
(integer) 3
```

### ZLEXCOUNT key min max

Returns the number of elements in the sorted set at key with a value between min and max.

**Return value**

int64: the number of elements in the specified score range.

**Example**

```
ledis> ZADD myzset 0 a 0 b 0 c 0 d 0 e 0 f 0 g
(integer) 7
ledis> ZLEXCOUNT myzset - [c
(integer) 3
```

### ZDUMP key

See [DUMP](#dump-key) for more information.

## Bitmap

### BGET key

Returns the whole binary data stored at `key`.

**Return value**

bulk: the raw value of key, or nil when key does not exist.

**Examples**

```
ledis> BMSETBIT flag 0 1 5 1 6 1
(integer) 3
ledis> BGET flag
a
```


### BGETBIT key offset

Returns the bit value at `offset` in the string value stored at `key`.
When *offset* beyond the data length, ot the target data is not exist, the bit value will be 0 always.

**Return value**

int64 : the bit value stored at offset.

**Examples**

```
ledis> BSETBIT flag 1024 1
(integer) 0
ledis> BGETBIT flag 0
(integer) 0
ledis> BGETBIT flag 1024
(integer) 1
ledis> BGETBIT flag 65535
(integer) 0
```


### BSETBIT key offset value

Sets or clear the bit at `offset` in the binary data sotred at `key`.
The bit is either set or cleared depending on `value`, which can be either `0` or `1`.
The *offset* argument is required to be qual to 0, and smaller than
2^23 (this means bitmap limits to 8MB).

**Return value**

int64 : the original bit value stored at offset.

**Examples**

```
ledis> BSETBIT flag 0 1
(integer) 0
ledis> BSETBIT flag 0 0
(integer) 1
ledis> BGETBIT flag 0 99
ERR invalid command param
```

### BMSETBIT key offset value [offset value ...]
Sets the given *offset* to their respective values.

**Return value**

int64 : The number of input *offset*

**Examples**

```
ledis> BMSETBIT flag 0 1 1 1 2 0 3 1
(integer) 4
ledis> BCOUNT flag
(integer) 3
```


### BOPT operation destkey key [key ...]
Perform a bitwise operation between multiple keys (containing string values) and store the result in the destination key.

**Return value**

Int64:
The size of the string stored in the destination key, that is equal to the size of the longest input string.
**Examples**

```
ledis> BMSETBIT a 0 1 2 1
(integer) 2
ledis> BMSETBIT b 1 1 
(integer) 1
ledis> BOPT AND res a b    
(integer) 3
ledis> BCOUNT res
(integer) 0
ledis> BOPT OR res2 a b
(integer) 3
ledis> BCOUNT res2
(integer) 3
ledis> BOPT XOR res3 a b
(integer) 3
ledis> BCOUNT res3
(integer) 3
```

### BCOUNT key [start end]

Count the number of set bits in a bitmap.

**Return value**

int64 : The number of bits set to 1.

**Examples**

```
ledis> BMSETBIT flag 0 1 5 1 6 1
(integer) 3
ledis> BGET flag
a
ledis> BCOUNT flag
(integer) 3
ledis> BCOUNT flag 0 0s
(integer) 1
ledis> BCOUNT flag 0 4
(integer) 1
ledis> BCOUNT flag 0 5
(integer) 2
ledis> BCOUNT flag 5 6
(integer) 2
```


### BEXPIRE key seconds

(refer to [EXPIRE](#expire-key-seconds) api for other types)


### BEXPIREAT key timestamp

(refer to [EXPIREAT](#expireat-key-timestamp) api for other types)


### BTTL key

(refer to [TTL](#ttl-key) api for other types)


### BPERSIST key

(refer to [PERSIST](#persist-key) api for other types)


### BXSCAN key [MATCH match] [COUNT count] 

Iterate Bitmap keys incrementally.

See [XSCAN](#xscan-key-match-match-count-count) for more information.

### BXREVSCAN key [MATCH match] [COUNT count] 

Reverse iterate Bitmap keys incrementally.

See [XREVSCAN](#xrevscan-key-match-match-count-count) for more information.


## Replication

### SLAVEOF host port [RESTART] [READONLY]

Changes the replication settings of a slave on the fly. If the server is already acting as slave, `SLAVEOF NO ONE` will turn off the replication and turn the server into master. `SLAVEOF NO ONE READONLY` will turn the server into master with readonly mode. 

If the server is already master, `SLAVEOF NO ONE READONLY` will force the server to readonly mode, and `SLAVEOF NO ONE` will disable readonly.

`SLAVEOF host port` will make the server a slave of another server listening at the specified host and port.

If a server is already a slave of a master, `SLAVEOF host port` will stop the replication against the old and start the synchronization against the new one, if RESTART is set, it will discard the old dataset, otherwise it will sync with LastLogID + 1. 


### FULLSYNC [NEW]

Inner command, starts a fullsync from the master set by SLAVEOF.

FULLSYNC will first try to sync all data from the master, save in local disk, then discard old dataset and load new one.

`FULLSYNC NEW` will generate a new snapshot and sync, otherwise it will use the latest existing snapshot if possible.

**Return value**

**Examples**


### SYNC logid

Inner command, syncs the new changed from master set by SLAVEOF with logid.

**Return value**

**Examples**

## Server

### PING
Returns PONG. This command is often used to test if a connection is still alive, or to measure latency.

**Return value**

String

**Examples**

```
ledis> PING
PONG
ledis> PING
dial tcp 127.0.0.1:6665: connection refused
ledis>
```

### ECHO message

Returns message.

**Return value**

bulk string reply

**Examples**

```
ledis> ECHO "hello"
hello
```

### SELECT index
Select the DB with having the specified zero-based numeric index. New connections always use DB `0`. Currently, We support `16` DBs(`0-15`).

**Return value**

Simple string reply

**Examples**

```
ledis> SELECT 2
OK
ledis> SELECT 15
OK
ledis> SELECT 16
ERR invalid db index 16
```

### FLUSHALL

Delete all the keys of all the existing databases and replication logs, not just the currently selected one. This command never fails.

Very dangerous to use!!!

### FLUSHDB

Delete all the keys of the currently selected DB. This command never fails.

Very dangerous to use!!!

### INFO [section]

Return information and statistic about the server in a format that is simple to parse by computers and easy to read by humans.

The optional parameter can be used to select a specific section of information. When no parameter is provided, all will return.

### TIME

The TIME command returns the current server time as a two items lists: a Unix timestamp and the amount of microseconds already elapsed in the current second

**Return value**

array: two elements, one is unix time in seconds, the other is microseconds.

### CONFIG REWRITE

Rewrites the config file the server was started with. 

**Unlike Redis rewrite, it will discard all comments in origin config file.**

**Return value**

String: OK or error msg.

### RESTORE key ttl value 

Create a key associated with a value that is obtained by deserializing the provided serialized value (obtained via DUMP, LDUMP, HDUMP, SDUMP, ZDUMP).

If ttl is 0 the key is created without any expire, otherwise the specified expire time (in milliseconds) is set. But you must know that now the checking ttl accuracy is second.

RESTORE checks the RDB version and data checksum. If they don't match an error is returned.

## Transaction

### BEGIN

Marks the start of a transaction block. Subsequent commands will be in a transaction context util using COMMIT or ROLLBACK.

You must known that `BEGIN` will block any other write operators before you `COMMIT` or `ROLLBACK`. Don't use long-time transaction.

**Return value**

Returns `OK` if the backend store engine in use supports transaction, otherwise, returns `Err`. 

**Examples**
```
ledis> BEGIN
OK
ledis> SET HELLO WORLD
OK
ledis> COMMIT
OK
```

### ROLLBACK

Discards all the changes of previously commands in a transaction and restores the connection state to normal.

**Return value**
Returns `OK` if in a transaction context, otherwise, `Err`

**Examples**
```
ledis> BEGIN
OK
ledis> SET HELLO WORLD
OK
ledis> GET HELLO
"WORLD"
ledis> ROLLBACK
OK
ledis> GET HELLO
(nil)
```

### COMMIT

Persists the changes of all the commands in a transaction and restores the connection state to normal.

**Return value**
Returns `OK` if in a transaction context, otherwise, `Err`

**Examples**
```
ledis> BEGIN
OK
ledis> SET HELLO WORLD
OK
ledis> GET HELLO
"WORLD"
ledis> COMMIT
OK
ledis> GET HELLO
"WORLD"
```

## Script

LedisDB's script is refer to Redis, you can see more [http://redis.io/commands/eval](http://redis.io/commands/eval)

You must notice that executing lua will block any other write operations.

### EVAL script numkeys key [key ...] arg [arg ...]

### EVALSHA sha1 numkeys key [key ...] arg [arg ...]

### SCRIPT LOAD script

### SCRIPT EXISTS script [script ...]

### SCRIPT FLUSH


Thanks [doctoc](http://doctoc.herokuapp.com/)
