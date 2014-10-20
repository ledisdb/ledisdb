
LedisDB is not Redis, so you can not use some Redis clients for LedisDB directly. 
But LedisDB uses Redis protocol for communication and many APIs are same as Redis, 
so you can easily write your own LedisDB client based on a Redis one.

Before you write a client, you must know some differences between LedisDB and Redis.

## Data Structure

LedisDB has no Strings data type but KV and Bitmap, any some Keys and Strings commands in Redis will only affect KV data, and "bit" commands affect Bitmap.

## Del

In Redis, `del` can delete all type data, like String, Hash, List, etc, but in LedisDB, `del` can only delete KV data. To delete other type data, you will use "clear" commands.

+ KV:     `del`, `mdel` 
+ Hash:   `hclear`, `mhclear` 
+ List:   `lclear`, `mlclear`
+ Set:    `sclear`, `msclear`  
+ Zset:   `zclear`, `mzclear`
+ Bitmap: `bclear`, `mbclear`

## Expire, Persist, and TTL

The same for Del.

+ KV:     `expire`, `persist`, `ttl` 
+ Hash:   `hexpire`, `hpersist`, `httl` 
+ List:   `lexpire`, `lpersist`, `lttl`
+ Set:    `sexpire`, `spersist`, `sttl`  
+ Zset:   `zexpire`, `zpersist`, `zttl`
+ Bitmap: `bexpire`, `bpersist`, `bttl`

## ZSet

ZSet only support int64 score, not double in Redis.

## Transaction

LedisDB supports ACID transaction using LMDB or BoltDB, maybe later it will support `multi`, `exec`, `discard`.

Transaction API:

+ `begin`
+ `commit`
+ `rollback`

## Scan

LedisDB supplies `xscan`, `xrevscan`, etc, to fetch data iteratively and reverse iteratively.

+ KV:     `xscan`, `xrevscan`
+ Hash:   `hxscan`, `hxrevscan`
+ List:   `lxscan`, `lxrevscan`
+ Set:    `sxscan` , `sxrevscan`
+ Zset:   `zxscan`, `zxrevscan`
+ Bitmap: `bxscan`, `bxrevscan`


Of course, LedisDB has not implemented all APIs in Redis, you can see full commands in commands.json, commands.doc or [wiki](https://github.com/siddontang/ledisdb/wiki/Commands).