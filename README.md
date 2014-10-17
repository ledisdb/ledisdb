# LedisDB

Ledisdb is a high performance NoSQL like Redis written by go. It supports some data structure like kv, list, hash, zset, bitmap,set.

LedisDB now supports multiple databases as backend to store data, you can test and choose the proper one for you.

## Features

+ Rich data structure: KV, List, Hash, ZSet, Bitmap, Set.
+ Stores lots of data, over the memory limit. 
+ Various backend database to use: LevelDB, goleveldb, LMDB, RocksDB, BoltDB, Memory.
+ Supports transaction using LMDB or BotlDB.
+ Supports lua scripting.
+ Supports expiration and ttl.
+ Supports using redis-cli directly.
+ Easy to embed in your own Go application. 
+ HTTP API support, json/bson/msgpack output.
+ Replication to guarantee data safe.
+ Supplies tools to load, dump, repair database. 

## Build and Install

Create a workspace and checkout ledisdb source

    mkdir $WORKSPACE
    cd $WORKSPACE
    git clone git@github.com:siddontang/ledisdb.git src/github.com/siddontang/ledisdb

    cd src/github.com/siddontang/ledisdb

    #install go dependences
    ./bootstrap.sh

    #set build and run environment 
    source dev.sh

    make
    make test

## Godep support

LedisDB supports building with [godep](https://github.com/tools/godep) which can manage LedisDB go dependence automatically.

## LevelDB support

+ Install leveldb and snappy.

    LedisDB supplies a simple script to install leveldb and snappy: 

        sudo sh tools/build_leveldb.sh

    It will default install leveldb at /usr/local/leveldb and snappy at /usr/local/snappy.

    LedisDB use the modified LevelDB for better performance, see [here](https://github.com/siddontang/ledisdb/wiki/leveldb-source-modification).

    You can use other LevelDB (like Hyper LevelDB, Basho LevelDB) instead easily, the only thing you may pay attention to is that the header files must be in `include/leveldb` folder not `include/hyperleveldb` or other. 

+ Set `LEVELDB_DIR` and `SNAPPY_DIR` to the actual install path in dev.sh.
+ `make clean && make` 

## RocksDB support

+ [Install rocksdb](https://github.com/facebook/rocksdb/blob/master/INSTALL.md)(`make shared_lib`) and snappy first.

    LedisDB has not supplied a simple script to install, maybe later.

+ Set `ROCKSDB_DIR` and `SNAPPY_DIR` to the actual install path in `dev.sh`.
+ `make clean && make` 


Because RocksDB API may change sometimes, LedisDB may not build successfully. Now LedisDB supports RocksDB version 3.5 or newest master branch.
    

## Choose store database

LedisDB now supports goleveldb, lmdb, leveldb, rocksdb, boltdb, memory. it will use goleveldb as default to store data if you don't set.

Choosing a store database to use is very simple, you have two ways:

+ Set in server config file

        db_name = "leveldb"

+ Set in command flag

        ledis-server -config=/etc/ledis.conf -db_name=leveldb

    Flag command set will overwrite config set.

## Lua support

+ Compile and install lua
+ Set `LUA_DIR` to the actual path in `dev.sh`
+ `make clean && make` 

## Configuration

LedisDB uses [toml](https://github.com/toml-lang/toml) as the configuration format. The basic configuration ```./etc/ledis.conf``` in LedisDB source may help you.

If you don't use a configuration, LedisDB will use the default for you.

## Server Example
    
    //set run environment if not
    source dev.sh

    ledis-server -config=/etc/ledis.conf

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

## Client

See [Clients](https://github.com/siddontang/ledisdb/wiki/Clients) to find or contribute LedisDB client.

## Links

+ [Official Website](http://ledisdb.com)
+ [GoDoc](https://godoc.org/github.com/siddontang/ledisdb)
+ [Server Commands](https://github.com/siddontang/ledisdb/wiki/Commands)

## Caveat

+ You must known that changing store database runtime is very dangerous, LedisDB will not guarantee the data validation if you do it.
+ Begin a transaction will block any other write operators before you call `commit` or `rollback`. Don't use long-time transaction. 
+ `pcall` and `xpcall` are not supported in lua, you can see the readme in [golua](https://github.com/aarzilli/golua).


## Thanks

Gmail: cenqichao@gmail.com

Gmail: chendahui007@gmail.com

Gmail: cppgohan@gmail.com

Gmail: tiaotiaoyly@gmail.com

Gmail: wyk4true@gmail.com


## Feedback

Gmail: siddontang@gmail.com
