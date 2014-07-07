local ledis = require "ledis"
local cjson = require "cjson"
local lds = ledis:new()

lds:set_timeout(1000)

-- connect
local ok, err =  lds:connect("127.0.0.1", "6380")
if not ok then
	ngx.say("failed to connect:", err)
	return
end


function cleanUp()
	lds:del("mykey", "key1", "key2", "key3", "non_exists_key")
	lds:hmclear("myhash", "myhash1", "myhash2")
	lds:lclear("mylist")
	lds:zmclear("myset", "myset1", "myset2")
	return
end

cleanUp()


ngx.say("======================= K/V =====================\n")

--[[KV]]--

-- decr
local res, err = lds:decr("mykey")
if not res then
	ngx.say("failed to decr:", err)
	return
end

ngx.say("DECR, should be: -1 <=> ", res)

lds:del("mykey")

-- decrby

local res, err = lds:decrby("mykey", 10)
if not res then
	ngx.say("failed to decrby:", err)
	return
end

ngx.say("DECRBY, should be: -10 <=> ", res)
lds:del("mykey")

-- del 

lds:set("key1", "foo")
lds:set("key2", "bar")
local res, err = lds:del("key1", "key2")
if not res then
	ngx.say("failed to del:", err)
	return
end

ngx.say("DEL, should be: 2 <=> 2")

--exists

lds:set("mykey", "foo")
res, err = lds:exists("mykey")
if not res then
	ngx.say("failed to exists: ", err)
	return
end

ngx.say("EXISTS, should be 1 <=>", res)
lds:del("mykey")

res, err = lds:exists("non_exists_key")
if not res then
	ngx.say("failed to exists: ", err)
	return
end
ngx.say("EXISTS, should be 0 <=>", res)
lds:del("non_exists_key")


-- get

lds:set("mykey", "foo")
res, err = lds:get("mykey")
if not res then
	ngx.say("failed to get: ", err)
	return
end

ngx.say("GET, should be foo <=> ", res)
lds:del("mykey")


-- getset

lds:set("mykey", "foo")
res, err = lds:getset("mykey", "bar")
if not res then
	ngx.say("failed to getset ", err)
	return
end

ngx.say("GETSET, should be foo <=> ", res)
res, err = lds:get("mykey")
ngx.say("GET, should be bar <=>", res)
lds:del("mykey")

-- incr

lds:set("mykey", "10")
res, err = lds:incr("mykey")
if not res then
	ngx.say("failed to incr ", err)
	return
end

ngx.say("INCR should be 11 <=>", res)
lds:del("mykey")

-- incrby

lds:set("mykey", "10")
res, err = lds:incrby("mykey", 10)
if not res then
	ngx.say("failed to incrby ", err)
	return
end

ngx.say("INCRBY should be 20 <=>", res)
lds:del("mykey")

-- mget
lds:set("key1", "foo")
lds:set("key2", "bar")
res, err = lds:mget("key1", "key2")
if not res then
	ngx.say("failed to mget ", err)
	return
end

ngx.say("MGET should be foobar <=>", res)
lds:del("key1", "key2")
		
-- mset

res, err = lds:mset("key1", "foo", "key2", "bar")
if not res then
	ngx.say("failed to command ", err)
	return
end

ngx.say("MSET should be OK <=>", res)
lds:del("key1", "key2")



-- set
ok, err = lds:set("mykey", "foo")
if not ok then
	ngx.say("failed to set: ", err)
	return
end

ngx.say("SET, should be  OK <=>", ok)
lds:del("mykey")

-- setnx
res, err = lds:setnx("mykey", "foo")
if not res then
	ngx.say("failed to setnx ", err)
	return
end
	
ngx.say("setnx should be 1 <=>", res)
res, err = lds:setnx("mykey", "foo")
ngx.say("setnx should be 0 <=>", res)
lds:del("mykey")

-- expire

lds:set("mykey", "foo")
res, err = lds:expire("mykey", 60)
if not res then
	ngx.say("failed to expire ", err)
	return
end

ngx.say("EXPIRE should be 1 <=> ", res)
lds:del("mykey")


-- expireat
lds:set("mykey", "foo")
res, err = lds:expire("mykey", 14366666666)
if not res then
	ngx.say("failed to expireat", err)
	return
end

ngx.say("EXPIREAT 1 <=>", res)
lds:del("mykey")

-- ttl

lds:set("mykey", "foo")
lds:expire("mykey", 100)
res, err = lds:ttl("mykey")
if not res then
	ngx.say("failed to ttl ", err)
	return
end

if not (0 < res and res <= 100) then
	ngx.say("failed to ttl")
	return
end
ngx.say("TTL ",  res)
lds:del("mykey")

-- persist

lds:set("mykey", "foo")
lds:expire("mykey", 100)
res, err = lds:persist("mykey")

if not res then
	ngx.say("failed to persist", err)
	return
end

ngx.say("PERSIST should be 1 <=>", res)
lds:del("mykey")

-- [[ HASH ]]

ngx.say("\n=================== HASH =====================\n")

-- hdel

res, err = lds:hset("myhash", "field", "foo")
if not res then
	ngx.say("failed to HDEL", err)
	return
end

ngx.say("HDEL should be 1 <=>", res)
lds:hclear("myhash")

-- hexists
lds:hset("myhash", "field", "foo")
res, err = lds:hexists("myhash", "field")
if not res then
	ngx.say("failed to HEXISTS", err)
	return
end

ngx.say("HEXISTS should be 1 <=>", res)
lds:hclear("myhash")

-- hget
lds:hset("myhash", "field", "foo")
res, err = lds:hget("myhash", "field")
if not res then
	ngx.say("failed to HGET ", err)
	return
end

ngx.say("HGET should be foo <=>", res)
lds:hclear("myhash")

-- hgetall
lds:hmset("myhash", "field1", "foo", "field2", "bar")
res, err = lds:hgetall("myhash")
if not res then
	ngx.say("failed to HGETALL ", err)
	return
end

ngx.say("HGETALL should be field1foofield2bar <=>", res)
lds:hclear("myhash")

-- hincrby
res, err = lds:hincrby("myhash", "field", 1)
if not res then
	ngx.say("failed to HINCRBY ", err)
	return
end

ngx.say("HINCRBY should be 1 <=>", res)
lds:hclear("myhash")

-- hkeys

lds:hmset("myhash", "field1", "foo", "field2", "bar")
res, err = lds:hkeys("myhash")
if not res then
	ngx.say("failed to HKEYS", err)
	return
end

ngx.say("HKEYS should be field1field2 <=> ", res)
lds:hclear("myhash")


-- hlen 

lds:hset("myhash", "field", "foo")
res, err = lds:hlen("myhash")
if not res then
	ngx.say("failed to HLEN ", err)
	return
end

ngx.say("HLEN should be 1 <=>", res)
lds:hclear("myhash")


-- hmget

lds:hmset("myhash", "field1", "foo", "field2", "bar")
res, err = lds:hmget("myhash", "field1", "field2")
if not res then
	ngx.say("failed to HMGET", err)
	return
end

ngx.say("HMGET 	should be foobar <=>", res)
lds:hclear("myhash")



-- hmset

res, err = lds:hmset("myhash", "field1", "foo", "field2", "bar")
if not res then
	ngx.say("failed to HMSET ", err)
	return
end

local l = lds:hlen("myhash")
if l == 2 then
	ngx.say("HMSET tested !")
else
	ngx.say("HMSET failed")
end

res, err = lds:hclear("myhash")


--hset
res, err = lds:hset("myhash", "field", "foo")
if not res then
	ngx.say("failed to HSET", err)
	return
end

ngx.say("HSET should be 1 <=> ", res)
lds:hclear("myhash")

--hvals
lds:hset("myhash", "field", "foo")
res, err = lds:hvals("myhash")
if not res then
	ngx.say("failed to HVALS", err)
	return
end

ngx.say("HVALS should  be foo <=>", res)
lds:hvals("myhash")

-- hclear

--FIXME: why 3?

lds:hset("myhash", "field", "foo")
res, err = lds:hclear("myhash")

if not res then
	ngx.say("failed to HCLEAR", err)
	return
end

ngx.say("HCLEAR should be 1 <=>", res)
lds:hclear("myhash")


-- hmclear

lds:hset("myhash1", "field1", "foo")
lds:hset("myhash2", "field2", "bar")
res, err = lds:hmclear("myhash1", "myhash2")
if not res then
	ngx.say("failed to HMCLEAR ", err)
	return
end

ngx.say("HMCLEAR should be 2 <=>", res)


-- hexpire
lds:hset("myhash", "field", "foo")
res, err = lds:hexpire("myhash", 100)
if not res then
	ngx.say("failed to HEXPIRE", err)
	return
end

ngx.say("HEXPIRE should be 1 <=>", res)
lds:hclear("myhash")


-- hexpireat

lds:hset("myhash", "field", "foo")
res, err = lds:hexpireat("myhash", 14366666666)
if not res then
	ngx.say("failed to HEXPIREAT", err)
	return
end

ngx.say("HEXPIREAT should be 1 <=>", res)
lds:hclear("myhash")


-- hpersist
lds:hset("myhash", "field", "foo")
lds:hexpire("myhash", 100)
res, err = lds:hpersist("myhash")

if not res then
	ngx.say("failed to hpersist", err)
	return
end

ngx.say("HPERSIST should be 1 <=>", res)
lds:hclear("myhash")


--httl
lds:hset("myhash", "field", "foo")
lds:hexpire("myhash", 100)
res, err = lds:httl("myhash")
if not res then
	ngx.say("failed to HTTL ", err)
	return
end

ngx.say("HTTL value: ", res)
lds:hclear("myhash")
		

ngx.say("\n================== LIST ====================\n")


-- [[ LIST ]]

-- lindex
lds:rpush("mylist", "one", "two", "three")
res, err = lds:lindex("mylist", 0)
if not res then
	ngx.say("failed to LINDEX ", err)
	return
end


ngx.say("LINDEX should be one <=>", res)
lds:lclear("mylist")


-- llen
lds:rpush("mylist", "foo", "bar")
res, err = lds:llen("mylist")
if not res then
	ngx.say("failed to LLEN ", err)
	return
end

ngx.say("LLEN should be 2 <=>", res)
lds:lclear("mylist")


-- lpop
lds:rpush("mylist", "one", "two", "three")
res, err = lds:lpop("mylist")
if not res then
	ngx.say("failed to LPOP ", err)
	return
end

ngx.say("LPOP should be one <=>", res)
lds:lclear("mylist")
		
-- lrange
lds:rpush("mylist", "one", "two", "three")
res, err = lds:lrange("mylist", 0, 0)
if not res then
	ngx.say("failed to one ", err)
	return
end
	
ngx.say("LRANGE should be one <=>", res)
lds:lclear("mylist")
		


-- lpush

res, err = lds:lpush("mylist", "one", "two")
if not res then
	ngx.say("failed to LPUSH ", err)
	return
end

ngx.say("LPUSH should be 2 <=>", res)
lds:lclear("mylist")
		

-- rpop

lds:rpush("mylist", "one", "two")
res, err = lds:rpop("mylist")
if not res then
	ngx.say("failed to RPOP ", err)
	return
end

ngx.say("RPOP should be two <=>", res)
lds:lclear("mylist")

-- rpush
res, err = lds:rpush("mylist", "one", "two")
if not res then
	ngx.say("failed to RPUSH ", err)
	return
end

ngx.say("RPUSH should be 2 <=>", res)
lds:lclear("mylist")

-- lclear
lds:rpush("mylist", "one", "two")
res, err = lds:lclear("mylist")
if not res then
	ngx.say("failed to LCLEAR ", err)
	return
end

ngx.say("LCLEAR should be 2 <=>", res)


-- lexpire
lds:rpush("mylist", "one")
res, err = lds:lexpire("mylist", 100)
if not res then
	ngx.say("failed to LEXPIRE", err)
	return
end

ngx.say("LEXPIRE should be 1 <=>", res)
lds:lclear("mylist")


-- lexpireat

lds:rpush("mylist", "one")
res, err = lds:lexpireat("mylist", 14366666666)
if not res then
	ngx.say("failed to LEXPIREAT", err)
	return
end

ngx.say("LEXPIREAT should be 1 <=>", res)
lds:lclear("mylist")


-- lpersist
lds:rpush("mylist", "one", "two")
lds:lexpire("mylist", 100)
res, err = lds:lpersist("mylist")

if not res then
	ngx.say("failed to lpersist", err)
	return
end

ngx.say("LPERSIST should be 1 <=>", res)
lds:hclear("mylist")


--lttl
lds:rpush("mylist", "field", "foo")
lds:lexpire("mylist", 100)
res, err = lds:lttl("mylist")
if not res then
	ngx.say("failed to LTTL ", err)
	return
end

ngx.say("LTTL value: ", res)
lds:lclear("mylist")


ngx.say("\n==================== ZSET =====================\n")

-- [[ ZSET ]]

-- zadd
res, err = lds:zadd("myset", 1, "one")
if not res then
	ngx.say("failed to ZADD ", err)
	return
end

ngx.say("ZADD should be 1 <=>", res)
lds:zclear("myset")

-- zcard
lds:zadd("myset", 1, "one", 2, "two")

res, err = lds:zcard("myset")
if not res then
	ngx.say("failed to ZCARD ", err)
	return
end

ngx.say("ZCARD should be 2 <=>", res)
lds:zclear("myset")
		

-- zcount
lds:zadd("myset", 1, "one", 2, "two")
res, err = lds:zcount("myset", "-inf", "+inf")
if not res then
	ngx.say("failed to ZCOUNT ", err)
	return
end

ngx.say("ZCOUNT should be 2 <=>", res)
lds:zclear("myset")
		
--zincrby
lds:zadd("myset", 1, "one")
res, err = lds:zincrby("myset", 2, "one")
if not res then
	ngx.say("failed to ZINCRBY ", err)
	return
end

ngx.say("ZINCRBY should be 3 <=>", res)
lds:zclear("myset")
		

--zrange
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zrange("myset", 0, -1, "WITHSCORES")
if not res then
	ngx.say("failed to ZRANGE ", err)
	return
end
	
ngx.say("ZRANGE should be one1two2three3<=>", res)
lds:zclear("myset")
		

--zrangebyscore
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zrangebyscore("myset", 1, 2)
if not res then
	ngx.say("failed to ZRANGEBYSCORE ", err)
	return
end

ngx.say("ZRANGEBYSCORE should be onetwo <=>", res)
lds:zclear("myset")


-- zrank
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zrank("myset", "three")
if not res then
	ngx.say("failed to ZRANK ", err)
	return
end

ngx.say("ZRANK should be 2 <=>", res)
lds:zclear("myset")

--zrem
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zrem("myset", "two", "three")
if not res then
	ngx.say("failed to ZREM ", err)
	return
end

ngx.say("ZREM should be 2 <=>", res)
lds:zclear("myset")


--zremrangebyrank
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err= lds:zremrangebyrank("myset", 0, 2)
if not res then
	ngx.say("failed to ZREMRANGEBYRANK ", err)
	return
end

ngx.say("ZREMRANGEBYRANK should be 3 <=>", res)
lds:zclear("myset")


--zremrangebyscore
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zremrangebyscore("myset", 0, 2)
if not res then
	ngx.say("failed to ZREMRANGEBYSCORE ", err)
	return
end

ngx.say("zremrangebyscore should be 2 <=>", res)
lds:zclear("myset")


-- zrevrange
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zrevrange("myset", 0, -1)
if not res then
	ngx.say("failed to ZREVRANGE ", err)
	return
end

ngx.say("ZREVRANGE should be threetwoone <=>", res)
lds:zclear("myset")




-- zrevrangebyscore
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zrevrangebyscore("myset", "+inf", "-inf")
if not res then
	ngx.say("failed to ZREVRANGEBYSCORE ", err)
	return
end

ngx.say("ZREVRANGEBYSCORE should be threetwoone <=>", res)
lds:zclear("myset")



-- zscore
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zscore("myset", "two")
if not res then
	ngx.say("failed to ZSCORE ", err)
	return
end

ngx.say("ZSCORE should be 2 <=>", res)
lds.zclear("myset")


-- zclear
lds:zadd("myset", 1, "one", 2, "two", 3, "three")
res, err = lds:zclear("myset")
if not res then
	ngx.say("failed to ZCLEAR ", err)
	return
end

ngx.say("ZCLEAR should be 3 <=>", res)


-- zmclear
lds:zadd("myset1", 1, "one", 2, "two", 3, "three")
lds:zadd("myset2", 1, "one", 2, "two", 3, "three")
res, err = lds:zmclear("myset1", "myset2")
if not res then
	ngx.say("failed to ZMCLEAR ", err)
	return
end

ngx.say("ZMCLEAR should be 2 <=>", res)

--zexpire

--zexpireat

--zpersist

--zttl


-- zexpire

lds:zadd("myset", 1, "one")
res, err = lds:zexpire("myset", 60)
if not res then
	ngx.say("failed to zexpire ", err)
	return
end

ngx.say("ZEXPIRE should be 1 <=> ", res)
lds:zclear("myset")


-- zexpireat
lds:zadd("myset", 1, "one")
res, err = lds:zexpire("myset", 14366666666)
if not res then
	ngx.say("failed to zexpireat", err)
	return
end

ngx.say("ZEXPIREAT 1 <=>", res)
lds:zclear("myset")

-- zttl

lds:zadd("myset", 1, "one")
lds:zexpire("myset", 100)
res, err = lds:zttl("myset")
if not res then
	ngx.say("failed to zttl ", err)
	return
end

if not (0 < res and res <= 100) then
	ngx.say("failed to zttl")
	return
end
ngx.say("ZTTL ",  res)
lds:zclear("myset")

-- zpersist

lds:zadd("myset", 1, "one")
lds:zexpire("myset", 100)
res, err = lds:zpersist("myset")

if not res then
	ngx.say("failed to zpersist", err)
	return
end

ngx.say("ZPERSIST should be 1 <=>", res)
lds:zclear("myset")


ngx.say("\n===================== SERVER INFO ==============\n")

-- [[ SERVER INFO ]]

-- ping
res, err = lds:ping()
if not res then
	ngx.say("failed to PING ", err)
	return
end

ngx.say("PING should be PONG <=>", res)

-- echo 
res, err = lds:echo("hello, lua")
if not res then
	ngx.say("failed to ECHO ", err)
	return
end

ngx.say("ECHO should be hello, lua <=>", res)


-- select

res, err = lds:select(5)
if not res then
	ngx.say("failed to SELECT ", err)
	return
end

ngx.say("SELECT should be OK <=>", res)


