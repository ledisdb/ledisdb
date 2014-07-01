## Total

ledisdb use redis protocol called RESP(REdis Serialization Protocol), [here](http://redis.io/topics/protocol).

ledisdb all commands return RESP fomrat and it will use int64 instead of  RESP integer, string instead of RESP simple string, bulk string instead of RESP bulk string, and array instead of RESP arrays below.

## KV 

### DECR key
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


### DECRBY key decrement

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

### DEL key [key ...]

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

### EXISTS key

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

### GET key

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

### GETSET key value

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

### INCR key

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

### INCRBY key increment

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

### MGET key [key ...]

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

### MSET key value [key value ...]

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

### SET key value

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

### SETNX key value

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

### EXPIRE key seconds

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

### EXPIREAT key timestamp

Set an expired unix timestamp on key. 

**Return value**

int64:

- 1 if the timeout was set
- 0 if key does not exist or the timeout could not be set

**Examples**

```
ledis> set mykey "Hello"
OK
ledis> expireat mykey 1293840000
(integer) 1
ledis> exists mykey
(integer) 0
```

### TTL key

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

### PERSIST key

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

### HDEL key field [field ...]

Removes the specified fiedls from the hash stored at key.

**Return value**

int64: the number of fields that were removed from the hash.

**Examples**

```
ledis> hset myhash field1 "foo"
(integer) 1
ledis> hdel myhash field1 field2
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
ledis> hset myhash field1 "foo"
(integer) 1
ledis> hexists myhash field1 
(integer) 1
ledis> hexists myhash field2
(integer) 0
```

### HGET key field

Returns the value associated with field in the hash stored at key.

**Return value**

bulk: the value associated with field, or nil.

**Examples**

```
ledis> hset myhash field1 "foo"
(integer) 1
ledis> hget myhash field1
"foo"
ledis> hget myhash field2
(nil)
```

### HGETALL key

Returns all fields and values of the hash stored at key.

**Return value**

array: list of fields and their values stored in the hash, or an empty list (using nil in ledis-cli)

**Examples**

```
ledis> hset myhash field1 "hello"
(integer) 1
ledis> hset myhash field2 "world"
(integer) 1
ledis> hgetall myhash
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
ledis> hincrby myhash field 1
(integer) 1
ledis> hget myhash field
"1"
ledis> hincrby myhash field 5
(integer) 6
ledis> hincrby myhash field -10
(integer) -4
```

### HKEYS key

Return all fields in the hash stored at key.

**Return value**

array: list of fields in the hash, or an empty list.

**Examples**

```
ledis> hset myhash field1 "hello"
(integer) 1
ledis> hset myhash field2 "world"
(integer) 1
ledis> hkeys myhash
1) "field1"
2) "field2"
```

### HLEN key

Returns the number of fields contained in the hash stored at key

**Return value**

int64: number of fields in the hash, or 0 when key does not exist.

**Examples**

```
ledis> hset myhash field1 "hello"
(integer) 1
ledis> hset myhash field2 "world"
(integer) 1
ledis> hlen myhash
(integer) 2
```

### HMGET key field [field ...]

Returns the values associated with the specified fields in the hash stored at key. If field does not exist in the hash, a nil value is returned.

**Return value**

array: list of values associated with the given fields.

**Examples**

```
ledis> hset myhash field1 "hello"
(integer) 1
ledis> hset myhash field2 "world"
(integer) 1
ledis> hmget myhash field1 field2 nofield
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
ledis> hmset myhash field1 "hello" field2 "world"
OK
ledis> hmget myhash field1 field2
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
ledis> hset myhash field1 "hello"
(integer) 1
ledis> hget myhash field1
"hello"
ledis> hset myhash field1 "world"
(integer) 0
ledis> hget myhash field1
"world"
```

### HVALS key

Returns all values in the hash stored at key.

**Return value**

array: list of values in the hash, or an empty list.

**Examples**

```
ledis> hset myhash field1 "hello"
(integer) 1
ledis> hset myhash field2 "world"
(integer) 1
ledis> hvals myhash
1) "hello"
2) "world"
```

### HCLEAR key 

Deletes the specified hash keys

**Return value**

int64: the number of fields in the hash stored at key

**Examples**

```
ledis> hmset myhash field1 "hello" field2 "world"
OK
ledis> hclear myhash
(integer) 2
```

### HMCLEAR key [key...]

Deletes the specified hash keys

**Return value**

int64: the number of input keys

**Examples**

```
ledis> hmset myhash field1 "hello" field2 "world"
OK
ledis> hmclear myhash
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
27.0.0.1:6666> hset myhash a  100
(integer) 1
127.0.0.1:6666> hget myhash a
100
127.0.0.1:6666> hexpire myhash 100
(integer) 1
127.0.0.1:6666> httl myhash
(integer) 94
127.0.0.1:6666> hpersist myhash
(integer) 1
127.0.0.1:6666> httl myhash
(integer) -1
127.0.0.1:6666> hexpire not_exists_key 100
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
127.0.0.1:6666> hset myhash a  100
(integer) 1
127.0.0.1:6666> hexpireat myhash 1404999999
(integer) 1
127.0.0.1:6666> httl myhash
(integer) 802475
127.0.0.1:6666> hexpireat not_exists_key  1404999999
(integer) 0
```

### HTTL key

Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, -1 returns.

**Return value**

int64: TTL in seconds

**Examples**

```
127.0.0.1:6666> hset myhash a  100
(integer) 1
127.0.0.1:6666> hexpireat myhash 1404999999
(integer) 1
127.0.0.1:6666> httl myhash
(integer) 802475
127.0.0.1:6666> httl not_set_timeout
(integer) -1
```

### HPERSIST key

Remove the expiration from a hash key, like persist similarly.
Remove the existing timeout on key

**Return value**

int64:

- 1 if the timeout was removed
- 0 if key does not exist or does not have an timeout

```
127.0.0.1:6666> hset myhash a  100
(integer) 1
127.0.0.1:6666> hexpireat myhash 1404999999
(integer) 1
127.0.0.1:6666> httl myhash
(integer) 802475
127.0.0.1:6666> hpersist myhash
(integer) 1
127.0.0.1:6666> httl myhash
(integer) -1
127.0.0.1:6666> hpersist not_exists_key
(integer) 0
```


## List

### LINDEX key index
Returns the element at index index in the list stored at key. The index is zero-based, so 0 means the first element, 1 the second element and so on. Negative indices can be used to designate elements starting at the tail of the list. Here, -1 means the last element, -2 means the penultimate and so forth.
When the value at key is not a list, an error is returned.

**Return value**

string: the requested element, or nil when index is out of range.

**Examples**

```
ledis > rpush a 1 2 3
(integer) 3
ledis > lindex a 0
1
ledis > lindex a 1
2
ledis > lindex a 2
3
ledis > lindex a 3
(nil)
ledis > lindex a -1
3
```

### LLEN key
Returns the length of the list stored at key. If key does not exist, it is interpreted as an empty list and 0 is returned. An error is returned when the value stored at key is not a list.

**Return value**

int64: the length of the list at key.

**Examples**

```
ledis > rpush a 'foo'
(integer) 1
ledis > rpush a 'bar'
(integer) 2
ledis > llen a
(integer) 2
```

### LPOP key
Removes and returns the first element of the list stored at key.

**Return value**

bulk: the value of the first element, or nil when key does not exist.

**Examples**

```
ledis > rpush a 'one'
(integer) 1
ledis > rpush a 'two'
(integer) 2
ledis > rpush a 'three'
(integer) 3
ledis > lpop a
'one'
```

### LRANGE key start stop
Returns the specified elements of the list stored at key. The offsets start and stop are zero-based indexes, with 0 being the first element of the list (the head of the list), 1 being the next element and so on.

**Return value**

array: list of elements in the specified range.

**Examples**

```
ledis > rpush a 'one' 'two' 'three'
(integer) 3
ledis > lrange a 0 0
1) "'one'"
ledis > lrange a -100 100
1) "'one'"
2) "'two'"
3) "'three'"
ledis > lrange a -3 2
1) "'one'"
2) "'two'"
3) "'three'"
ledis > lrange a 0 -1
(empty list or set)
```
### LPUSH key value [value ...]
Insert all the specified values at the head of the list stored at key. If key does not exist, it is created as empty list before performing the push operations. When key holds a value that is not a list, an error is returned.

**Return value**

int64: the length of the list after the push operations.

**Examples**

```
ledis > lpush a 1
(integer) 1
ledis > lpush a 2
(integer) 2
ledis > lrange a 0 2
1) "2"
2) "1"
```

### RPOP key
Removes and returns the last element of the list stored at key.

**Return value**

bulk: the value of the last element, or nil when key does not exist.

**Examples**

```
edis > rpush a 1
(integer) 1
ledis > rpush a 2
(integer) 2
ledis > rpush a 3
(integer) 3
ledis > rpop a
3
ledis > lrange a 0 3
1) "1"
2) "2"
```

### RPUSH key value [value ...]
Insert all the specified values at the tail of the list stored at key. If key does not exist, it is created as empty list before performing the push operation. When key holds a value that is not a list, an error is returned.

**Return value**

int64: the length of the list after the push operation.

**Examples**

```
ledis >  rpush a 'hello'
(integer) 1
ledis > rpush a 'world'
(integer) 2
ledis > lrange a 0 2
1) "'hello'"
2) "'world'"
```

### LCLEAR key
Deletes the specified list key

**Return value**

int64: the number of values in the list stored at key

**Examples**

```
ledis > rpush a 1 2 3
(integer) 3
ledis > llen a
(integer) 3
ledis > lclear a
(integer) 3
ledis > llen a
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
ledis > rpush a 1
(integer) 1
ledis > lexpire a 100
(integer) 1
ledis > lttl a
(integer) 96
ledis > lpersist a
(integer) 1
ledis > lttl a
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
ledis > rpush a 1
(integer) 1
ledis > lexpireat a 1404140183
(integer) 1
ledis > lttl a
(integer) 570
ledis > lpersist a
(integer) 1
ledis > lttl a
(integer) -1
ledis >
```

### LTTL key
Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, -1 returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis > rpush a 1
(integer) 1
ledis > lexpireat a 1404140183
(integer) 1
ledis > lttl a
(integer) 570
ledis > lpersist a
(integer) 1
ledis > lttl a
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
ledis > rpush a 1
(integer) 1
ledis > lexpireat a 1404140183
(integer) 1
ledis > lttl a
(integer) 570
ledis > lpersist a
(integer) 1
ledis > lttl a
(integer) -1
ledis > lpersist b
(integer) 0
```


## ZSet

### ZADD key score member [score member ...]
Adds all the specified members with the specified scores to the sorted set stored at key. It is possible to specify multiple score / member pairs. If a specified member is already a member of the sorted set, the score is updated and the element reinserted at the right position to ensure the correct ordering.

If key does not exist, a new sorted set with the specified members as sole members is created, like if the sorted set was empty. If the key exists but does not hold a sorted set, an error is returned.

The score values should be the string representation of a double precision floating point number. +inf and -inf values are valid values as well.

**Return value**

int64, specifically:

The number of elements added to the sorted sets, ** not ** including elements already existing for which the score was updated.


**Examples**

```
ledis > zadd myset 1 'one'
(integer) 1
ledis > zadd myset 1 'uno'
(integer) 1
ledis > zadd myset 2 'two' 3 'three'
(integer) 2
ledis > zrange myset 0 -1 withscores
1) "'one'"
2) "1"
3) "'uno'"
4) "1"
5) "'two'"
6) "2"
7) "'three'"
8) "3"
```

### ZCARD key
Returns the sorted set cardinality (number of elements) of the sorted set stored at key.

**Return value**

int64: the cardinality (number of elements) of the sorted set, or 0 if key does not exist.

**Examples**

```
edis > zadd myset 1 'one'
(integer) 1
ledis > zadd myset 1 'uno'
(integer) 1
ledis > zadd myset 2 'two' 3 'three'
(integer) 2
ledis > zrange myset 0 -1 withscores
1) "'one'"
2) "1"
3) "'uno'"
4) "1"
5) "'two'"
6) "2"
7) "'three'"
8) "3"
ledis > zcard myset
(integer) 4
```

### ZCOUNT key min max
Returns the number of elements in the sorted set at key with a score between min and max.
The min and max arguments have the same semantic as described for ZRANGEBYSCORE.

**Return value**

int64: the number of elements in the specified score range.

**Examples**

```
ledis > zadd myset 1 'one'
(integer) 1
ledis > zadd myset 1 'uno'
(integer) 1
ledis > zadd myset 2 'two' 3 'three'
(integer) 2
ledis > zrange myset 0 -1 withscores
1) "'one'"
2) "1"
3) "'uno'"
4) "1"
5) "'two'"
6) "2"
7) "'three'"
8) "3"
ledis > zcount myset -inf +inf
(integer) 4
ledis > zcount myset (1 3
(integer) 2
```

### ZINCRBY

Increments the score of member in the sorted set stored at key by increment. If member does not exist in the sorted set, it is added with increment as its score (as if its previous score was 0.0). If key does not exist, a new sorted set with the specified member as its sole member is created.
An error is returned when key exists but does not hold a sorted set.
The score value should be the string representation of a numeric value, and accepts double precision floating point numbers. It is possible to provide a negative value to decrement the score.

**Return value**

bulk: the new score of member (a double precision floating point number), represented as string.

**Examples**

```
ledis > zadd myset 1 'one'
(integer) 1
ledis > zadd myset 2 'two'
(integer) 1
ledis > zincrby myset 2 'one'
3
ledis > zrange myset 0 -1 withscores
1) "'two'"
2) "2"
3) "'one'"
4) "3"
```

### ZRANGE key start stop [WITHSCORES]
Returns the specified range of elements in the sorted set stored at key. The elements are considered to be ordered from the lowest to the highest score. Lexicographical order is used for elements with equal score.

**Return value**

array: list of elements in the specified range (optionally with their scores).

**Examples**

```
ledis > zadd myset 1 'one'
(integer) 1
ledis > zadd myset 2 'two'
(integer) 1
ledis > zadd myset 3 'three'
(integer) 1
ledis > zrange myset 0 -1
1) "'one'"
2) "'two'"
3) "'three'"
ledis > zrange myset 2 3
1) "'three'"
ledis > zrange myset -2 -1
1) "'two'"
2) "'three'"
```

### ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT off]
**Return value**

**Examples**
### ZRANK
**Return value**

**Examples**
### ZREM
**Return value**

**Examples**
### ZREMRANGEBYRANK
**Return value**

**Examples**
### ZREMRANGEBYSCORE
**Return value**

**Examples**
### ZREVRANGE
**Return value**

**Examples**
### ZREVRANGEBYSCORE
**Return value**

**Examples**

### ZSCORE key member
Returns the score of member in the sorted set at key.
If member does not exist in the sorted set, or key does not exist, nil is returned.

**Return value**

bulk: the score of member (a double precision floating point number), represented as string.

**Examples**

```
ledis > zadd myset 1 'one'
(integer) 1
ledis > zscore myset 'one'
1
```

### ZCLEAR key
Delete the specified  key

**Return value**

int64: the number of members in the zset stored at key

**Examples**

```
ledis > zadd myset 1 'one'
(integer) 1
ledis > zadd myset 2 'two'
(integer) 1
ledis > zadd myset 3 'three'
(integer) 1
ledis > zrange myset 0 -1
1) "'one'"
2) "'two'"
3) "'three'
ledis > zclear myset
(integer) 3
```

### ZMCLEAR key [key ...]
Delte multiple keys one time.

**Return value**

int64: the number of input keys

**Examples**

```
ledis > zadd myset1 1 'one'
(integer) 1
ledis > zadd myset2 2 'two'
(integer) 1
ledis > zmclear myset1 myset2
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
ledis > zadd myset 1 'one'
(integer) 1
ledis > zexpire myset 100
(integer) 1
ledis > zttl myset
(integer) 97
ledis > zpersist myset
(integer) 1
ledis > zttl mset
(integer) -1
ledis > zexpire myset1 100
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
ledis > zadd myset 1 'one'
(integer) 1
ledis > zexpireat myset 1404149999
(integer) 1
ledis > zttl myset
(integer) 7155
ledis > zpersist myset
(integer) 1
ledis > zttl mset
(integer) -1
ledis > zexpireat myset1 1404149999
(integer) 0
```


### ZTTL key
Returns the remaining time to live of a key that has a timeout. If the key was not set a timeout, -1 returns.

**Return value**

int64: TTL in seconds

**Examples**

```
ledis > zadd myset 1 'one'
(integer) 1
ledis > zexpire myset 100
(integer) 1
ledis > zttl myset
(integer) 97
ledis > zttl myset2
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
ledis > zadd myset 1 'one'
(integer) 1
ledis > zexpire myset 100
(integer) 1
ledis > zttl myset
(integer) 97
ledis > zpersist myset
(integer) 1
ledis > zttl mset
(integer) -1
```


## Replication

### SLAVEOF
**Return value**

**Examples**
### FULLSYNC
**Return value**

**Examples**
### SYNC
**Return value**

**Examples**



## Server

### PING
Returns PONG. This command is often used to test if a connection is still alive, or to measure latency.

**Return value**

String

**Examples**

```
ledis > ping
PONG
ledis > ping
dial tcp 127.0.0.1:6665: connection refused
ledis >
```

### ECHO message

Returns message.

**Return value**

bulk string reply

**Examples**

```
ledis > echo "hello"
hello
```

### SELECT index
Select the DB with having the specified zero-based numeric index. New connections always use DB 0. Currently, We support 16 dbs(0-15).

**Return value**

Simple string reply

**Examples**

```
ledis > select 2
OK
ledis > select 15
OK
ledis > select 16
ERR invalid db index 16
```
