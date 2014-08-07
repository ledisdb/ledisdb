# LedisDB

Ledisdb is a high performance NoSQL like Redis written by go. It supports some advanced data structure like kv, list, hash, zset, bitmap, and may be alternative for Redis.

LedisDB now supports multi databases as backend to store data, you can test and choose the proper one for you.

## Features

+ Rich advanced data structure: KV, List, Hash, ZSet, Bitmap.
+ Stores lots of data, over the memory limit. 
+ Various backend database to use: LevelDB, goleveldb, LMDB, RocksDB, BoltDB.  
+ Supports expiration and ttl.
+ Redis clients, like redis-cli, are supported directly.
+ Multi client API supports, including Go, Python, Lua(Openresty). 
+ Easy to embed in your own Go application. 
+ Restful API support, json/bson/msgpack output.
+ Replication to guarantee data safe.
+ Supplies tools to load, dump, repair database. 

## Build and Install

Create a workspace and checkout ledisdb source

    mkdir $WORKSPACE
    cd $WORKSPACE
    git clone git@github.com:siddontang/ledisdb.git src/github.com/siddontang/ledisdb

    cd src/github.com/siddontang/ledisdb

    #set build and run environment 
    source dev.sh

    make
    make test


## LevelDB support

+ Install leveldb and snappy.

    LedisDB supplies a simple script to install leveldb and snappy: 

        sh build_tool/build_leveldb.sh

    It will default install leveldb at /usr/local/leveldb and snappy at /usr/local/snappy.

    LedisDB use the modified LevelDB for better performance, see [here](https://github.com/siddontang/ledisdb/wiki/leveldb-source-modification).

+ Set ```LEVELDB_DIR``` and ```SNAPPY_DIR``` to the actual install path in dev.sh.
+ ```make```

## RocksDB support

+ Install rocksdb(shared_lib) and snappy first.

    LedisDB has not supplied a simple script to install, maybe later.

+ Set ```ROCKSDB_DIR``` and ```SNAPPY_DIR``` to the actual install path in dev.sh.
+ ```make```

## Choose store database

LedisDB now supports goleveldb, lmdb, leveldb, rocksdb, boltdb, it will choose goleveldb as default to store data if you not set.

Choosing a store database to use is very simple, you have two ways:

+ Set in server config file

        db_name = "leveldb"

+ Set in command flag

        ledis-server -config=/etc/ledis.toml -db_name=leveldb

    Flag command set will overwrite config set.

**Caveat**

You must known that changing store database runtime is very dangerous, LedisDB will not guarantee the data validation if you do it.

## Server Example
    
    //set run environment if not
    source dev.sh

    ledis-server -config=/etc/ledis.toml

    //another shell
    ledis-cli -p 6380
    
    ledis 127.0.0.1:6380> set a 1
    OK
    ledis 127.0.0.1:6380> get a
    "1"

    //use curl
    curl http://127.0.0.1:11181/SET/hello/world
    → {"SET":[true,"OK"]}

    curl http://127.0.0.1:11181/0/GET/hello?type=json
    → {"GET":"world"}


## Package Example
    
    import "github.com/siddontang/ledisdb/ledis"
    l, _ := ledis.Open(cfg)
    db, _ := l.Select(0)

    db.Set(key, value)

    db.Get(key)


## Replication Example

Set slaveof in config or dynamiclly

    ledis-cli -p 6381 

    ledis 127.0.0.1:6381> slaveof 127.0.0.1 6380
    OK

## Benchmark

See [benchmark](https://github.com/siddontang/ledisdb/wiki/Benchmark) for more.

## Todo

See [Issues todo](https://github.com/siddontang/ledisdb/issues?labels=todo&page=1&state=open)


## Links

+ [Official Website](http://ledisdb.com)
+ [Author's Chinese Blog](http://blog.csdn.net/siddontang/article/category/2264003)
+ [GoDoc](https://godoc.org/github.com/siddontang/ledisdb)
+ [Server Commands](https://github.com/siddontang/ledisdb/wiki/Commands)


## Thanks

Gmail: cenqichao@gmail.com

Gmail: chendahui007@gmail.com

Gmail: cppgohan@gmail.com

Gmail: tiaotiaoyly@gmail.com

Gmail: wyk4true@gmail.com


## Feedback

Gmail: siddontang@gmail.com
