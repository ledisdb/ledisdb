## Summary

ledisdb use redis protocol called RESP(REdis Serialization Protocol), [here](http://redis.io/topics/protocol).

ledisdb all commands return RESP fomrat and it will use `int64` instead of  `RESP integer`, `string` instead of `RESP simple string`, `bulk string` instead of `RESP bulk string`, and `array` instead of `RESP arrays` below.

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
	- [EXPIRE key seconds](#expire-key-seconds)
	- [EXPIREAT key timestamp](#expireat-key-timestamp)
	- [TTL key](#ttl-key)
	- [PERSIST key](#persist-key)
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
	- [HMCLEAR key [key...]](#hmclear-key-key)
	- [HEXPIRE key seconds](#hexpire-key-seconds)
	- [HEXPIREAT key timestamp](#hexpireat-key-timestamp)
	- [HTTL key](#httl-key)
	- [HPERSIST key](#hpersist-key)
- [List](#list)
	- [LINDEX key index](#lindex-key-index)
	- [LLEN key](#llen-key)
	- [LPOP key](#lpop-key)
	- [LRANGE key start stop](#lrange-key-start-stop)
	- [LPUSH key value [value ...]](#lpush-key-value-value-)
	- [RPOP key](#rpop-keuser-content-y)
	- [RPUSH key value [value ...]](#rpush-key-value-value-)
	- [LCLEAR key](#lclear-key)
	- [LEXPIRE key seconds](#lexpire-key-seconds)
	- [LEXPIREAT key timestamp](#lexpireat-key-timestamp)
	- [LTTL key](#lttl-key)
	- [LPERSIST key](#lpersist-key)
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
	- [ZREVRANGEBYSCORE  key max min [WITHSCORES] [LIMIT offset count]](#zrevrangebyscore--key-max-min-withscores-limit-offset-count)
	- [ZSCORE key member](#zscore-key-member)
	- [ZCLEAR key](#zclear-key)
	- [ZMCLEAR key [key ...]](#zmclear-key-key-)
	- [ZEXPIRE key seconds](#zexpire-key-seconds)
	- [ZEXPIREAT key timestamp](#zexpireat-key-timestamp)
	- [ZTTL key](#zttl-key)
	- [ZPERSIST key](#zpersist-key)
- [Replication](#replication)
	- [SLAVEOF host port](#slaveof-host-port)
	- [FULLSYNC](#fullsync)
	- [SYNC index offset](#sync-index-offset)
- [Server](#server)
	- [PING](#ping)
	- [ECHO message](#echo-message)
	- [SELECT index](#select-index)


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

Deletes the specified hash keys

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


## List

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
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZADD myset 1 'uno'
(integer) 1
ledis> ZADD myset 2 'two' 3 'three'
(integer) 2
ledis> ZRANGE myset 0 -1 WITHSCORES
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
edis > ZADD myset 1 'one'
(integer) 1
ledis> ZADD myset 1 'uno'
(integer) 1
ledis> ZADD myset 2 'two' 3 'three'
(integer) 2
ledis> ZRANGE myset 0 -1 WITHSCORES
1) "one"
2) "1"
3) "uno"
4) "1"
5) "two"
6) "2"
7) "three"
8) "3"
ledis> zcard myset
(integer) 4
```

### ZCOUNT key min max
Returns the number of elements in the sorted set at key with a score between `min` and `max`.
The `min` and `max` arguments have the same semantic as described for `ZRANGEBYSCORE`.

**Return value**

int64: the number of elements in the specified score range.

**Examples**

```
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZADD myset 1 'uno'
(integer) 1
ledis> ZADD myset 2 'two' 3 'three'
(integer) 2
ledis> ZRANGE myset 0 -1 WITHSCORES
1) "one"
2) "1"
3) "uno"
4) "1"
5) "two"
6) "2"
7) "three"
8) "3"
ledis> zcount myset -inf +inf
(integer) 4
ledis> zcount myset (1 3
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
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZADD myset 2 'two'
(integer) 1
ledis> ZINCRBY myset 2 'one'
3
ledis> ZRANGE myset 0 -1 WITHSCORES
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
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZADD myset 2 'two'
(integer) 1
ledis> ZADD myset 3 'three'
(integer) 1
ledis> ZRANGE myset 0 -1
1) "one"
2) "two"
3) "three"
ledis> ZRANGE myset 2 3
1) "three"
ledis> ZRANGE myset -2 -1
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
ledis> ZADD myset 1 one 2 two 3 three 4 four
(integer) 3
ledis> ZRANGE myset 0 -1
1) "one"
2) "two"
3) "three"
4) "four"
ledis> ZREM myset three
(integer) 1
ledis> ZREM myset one four three
(integer) 2
```

### ZREMRANGEBYRANK key start stop
Removes all elements in the sorted set stored at key with rank between start and stop. Both start and stop are 0 -based indexes with 0 being the element with the lowest score. These indexes can be negative numbers, where they indicate offsets starting at the element with the highest score. For example: -1 is the element with the highest score, -2 the element with the second highest score and so forth.

**Return value**

int64: the number of elements removed.

**Examples**

```
ledis> ZADD myset 1 one 2 two 3 three 4 four
(integer) 3
ledis> ZREMRANGEBYRANK myset 0 2
(integer) 3
ledis> ZRANGE myset 0 -1 WITHSCORES
1) "four"
2) "4"
```


### ZREMRANGEBYSCORE key min max
Removes all elements in the sorted set stored at key with a score between `min` and `max` (inclusive). `Min` and `max` can be exclusive, following the syntax of `ZRANGEBYSCORE`.

**Return value**

int64: the number of elements removed.

**Examples**

```
ledis> ZADD myset 1 one 2 two 3 three 4 four
(integer) 4
ledis> ZREMRANGEBYSCORE myset -inf (2
(integer) 1
ledis> ZRANGE myset 0 -1 WITHSCORES
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
ledis> ZADD myset 1 one 2 two 3 three 4 four
(integer) 4
ledis> ZREVRANGE myset 0 -1
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
ledis>  ZADD myset 1 one 2 two 3 three 4 four
(integer) 4
ledis> ZREVRANGEBYSCORE myset +inf -inf
1) "four"
2) "three"
3) "two"
4) "one"
ledis> ZREVRANGEBYSCORE myset 2 1
1) "two"
2) "one"
ledis> ZREVRANGEBYSCORE myset 2 (1
1) "two"
ledis> ZREVRANGEBYSCORE myset (2 (1
(empty list or set)
ledis> ZREVRANGEBYSCORE myset +inf -inf WITHSCORES LIMIT 1 2
1) "three"
2) "3"
3) "two"
4) "2"
```

### ZSCORE key member
Returns the score of member in the sorted set at key.
If member does not exist in the sorted set, or key does not exist, `nil` is returned.

**Return value**

bulk: the score of member (an `int64` number), represented as string.

**Examples**

```
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZSCORE myset 'one'
1
```

### ZCLEAR key
Delete the specified  key

**Return value**

int64: the number of members in the zset stored at key

**Examples**

```
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZADD myset 2 'two'
(integer) 1
ledis> ZADD myset 3 'three'
(integer) 1
ledis> ZRANGE myset 0 -1
1) "one"
2) "two"
3) "three"
ledis> ZCLEAR myset
(integer) 3
```

### ZMCLEAR key [key ...]
Delte multiple keys one time.

**Return value**

int64: the number of input keys

**Examples**

```
ledis> ZADD myset1 1 'one'
(integer) 1
ledis> ZADD myset2 2 'two'
(integer) 1
ledis> ZMCLEAR myset1 myset2
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
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZEXPIRE myset 100
(integer) 1
ledis> ZTTL myset
(integer) 97
ledis> ZPERSIST myset
(integer) 1
ledis> ZTTL mset
(integer) -1
ledis> ZEXPIRE myset1 100
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
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZEXPIREAT myset 1404149999
(integer) 1
ledis> ZTTL myset
(integer) 7155
ledis> ZPERSIST myset
(integer) 1
ledis> ZTTL mset
(integer) -1
ledis> ZEXPIREAT myset1 1404149999
(integer) 0
```


### ZTTL key
Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, `-1` returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis> zadd myset 1 'one'
(integer) 1
ledis> zexpire myset 100
(integer) 1
ledis> zttl myset
(integer) 97
ledis> zttl myset2
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
ledis> ZADD myset 1 'one'
(integer) 1
ledis> ZEXPIRE myset 100
(integer) 1
ledis> ZTTL myset
(integer) 97
ledis> ZPERSIST myset
(integer) 1
ledis> ZTTL mset
(integer) -1
```


## Replication

### SLAVEOF host port

Changes the replication settings of a slave on the fly. If the server is already acting as slave, SLAVEOF NO ONE will turn off the replication.

SLAVEOF host port will make the server a slave of another server listening at the specified host and port.

If a server is already a slave of a master, SLAVEOF host port will stop the replication against the old and start the synchronization against the new one, discarding the old dataset.


### FULLSYNC

Inner command, starts a fullsync from the master set by SLAVEOF.

FULLSYNC will first try to sync all data from the master, save in local disk, then discard old dataset and load new one.

**Return value**

**Examples**


### SYNC index offset

Inner command, syncs the new changed from master set by SLAVEOF at offset in binlog.index file.

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

Thanks [doctoc](http://doctoc.herokuapp.com/)
