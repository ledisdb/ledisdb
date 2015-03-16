
LedisDB is not Redis, so you can not use some Redis clients for LedisDB directly. 
But LedisDB uses Redis protocol for communication and many APIs are same as Redis, 
so you can easily write your own LedisDB client based on a Redis one.

Before you write a client, you must know some differences between LedisDB and Redis.

## Del

In Redis, `del` can delete all type data, like String, Hash, List, etc, but in LedisDB, `del` can only delete KV data. To delete other type data, you will use "clear" commands.

+ KV:     `del`, `mdel` 
+ Hash:   `hclear`, `hmclear` 
+ List:   `lclear`, `lmclear`
+ Set:    `sclear`, `smclear`  
+ ZSet:   `zclear`, `zmclear`

## Expire, Persist, and TTL

The same for Del.

+ KV:     `expire`, `persist`, `ttl` 
+ Hash:   `hexpire`, `hpersist`, `httl` 
+ List:   `lexpire`, `lpersist`, `lttl`
+ Set:    `sexpire`, `spersist`, `sttl`  
+ Zset:   `zexpire`, `zpersist`, `zttl`

## ZSet

ZSet only support int64 score, not double in Redis.


## Scan

LedisDB supplies `xscan`, `xhscan`, `xsscan`, `xzscan` to fetch data iteratively and reverse iteratively.

```
XSCAN type cursor [MATCH match] [COUNT count]
XHSCAN key cursor [MATCH match] [COUNT count]
XSSCAN key cursor [MATCH match] [COUNT count]
XZSCAN key cursor [MATCH match] [COUNT count]
```

## DUMP

+ KV: `dump`
+ Hash: `hdump`
+ List: `ldump`
+ Set: `sdump`
+ ZSet: `zdump`

LedisDB supports `dump` to serialize the value with key, the data format is the same as Redis, so you can use it in Redis and vice versa. 

Of course, LedisDB has not implemented all APIs in Redis, you can see full commands in commands.json, commands.doc or [wiki](https://github.com/siddontang/ledisdb/wiki/Commands).