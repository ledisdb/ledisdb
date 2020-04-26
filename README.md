# LedisDB 

[![Build Status](https://travis-ci.org/ledisdb/ledisdb.svg?branch=develop)](https://travis-ci.org/siddontang/ledisdb) [![codecov](https://codecov.io/gh/ledisdb/ledisdb/branch/master/graph/badge.svg)](https://codecov.io/gh/ledisdb/ledisdb)

Ledisdb is a high-performance NoSQL database, similar to Redis, written in [Go](http://golang.org/). It supports many data structures including kv, list, hash, zset, set.

LedisDB now supports multiple different databases as backends.

### **You must run `ledis-upgrade-ttl` before using LedisDB version 0.4, I fixed a very serious bug for key expiration and TTL.**

## Features

+ Rich data structure: KV, List, Hash, ZSet, Set.
+ Data storage is not limited by RAM.
+ Various backends supported: LevelDB, goleveldb, RocksDB, RAM.
+ Supports Lua scripting.
+ Supports expiration and TTL.
+ Can be managed via redis-cli.
+ Easy to embed in your own Go application. 
+ HTTP API support, JSON/BSON/msgpack output.
+ Replication to guarantee data safety.
+ Supplies tools to load, dump, and repair database. 
+ Supports cluster, use [xcodis](https://github.com/siddontang/xcodis)
+ Authentication (though, not via http)

## Build from source

Create a workspace and checkout ledisdb source

    git clone git@github.com:ledisdb/ledisdb.git
    cd ledisdb

    #set build and run environment 
    source dev.sh

    make
    make test

Then you will find all the binary build on `./bin` directory.

## LevelDB support

+ Install leveldb and snappy.

    LedisDB supplies a simple script to install leveldb and snappy: 

        sudo sh tools/build_leveldb.sh

    It will install leveldb at /usr/local/leveldb and snappy at /usr/local/snappy by default.

    LedisDB uses the modified LevelDB for better performance. [Details.](https://github.com/ledisdb/ledisdb/wiki/leveldb-source-modification)

    You can easily use other LevelDB versions (like Hyper LevelDB or Basho LevelDB) instead, as long as the header files are in `include/leveldb`, not `include/hyperleveldb` or any other location.

+ Set `LEVELDB_DIR` and `SNAPPY_DIR` to the actual install path in dev.sh.
+ `make clean && make` 

## RocksDB support 

+ [Install rocksdb(5.1+)](https://github.com/facebook/rocksdb/blob/master/INSTALL.md)(`make shared_lib`) and snappy first.

    LedisDB has not yet supplied a simple script to install.

+ Set `ROCKSDB_DIR` and `SNAPPY_DIR` to the actual install path in `dev.sh`.
+ `make clean && make` 


If the RocksDB API changes, LedisDB may not build successfully. LedisDB currently supports RocksDB version 5.1 or later.
    

## Choose store database

LedisDB now supports goleveldb, leveldb, rocksdb, and RAM. It will use goleveldb by default. 

Choosing a store database to use is very simple.

+ Set in server config file

        db_name = "leveldb"

+ Set in command flag

        ledis-server -config=/etc/ledis.conf -db_name=leveldb

    Flag command set will overwrite config setting.

## Lua support

Lua is supported using [gopher-lua](https://github.com/yuin/gopher-lua), a Lua VM, completely written in Go.

## Configuration

LedisDB uses [toml](https://github.com/toml-lang/toml) as the configuration format. The basic configuration ```./etc/ledis.conf``` in LedisDB source may help you.

If you don't use a configuration, LedisDB will use the default for you.

## Server Example
    
    //set run environment if not
    source dev.sh

    ./bin/ledis-server -config=/etc/ledis.conf

    //another shell
    ./bin/ledis-cli -p 6380
    
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
    
    import (
      lediscfg "github.com/ledisdb/ledisdb/config"
      "github.com/ledisdb/ledisdb/ledis"
    )

    # Use Ledis's default config
    cfg := lediscfg.NewConfigDefault()
    l, _ := ledis.Open(cfg)
    db, _ := l.Select(0)

    db.Set(key, value)

    db.Get(key)


## Replication Example

Set slaveof in config or dynamiclly

    ledis-cli -p 6381 

    ledis 127.0.0.1:6381> slaveof 127.0.0.1 6380
    OK

## Cluster support

LedisDB uses a proxy named [xcodis](https://github.com/siddontang/xcodis) to support cluster.

## CONTRIBUTING

See [CONTRIBUTING.md] .

## Benchmark

See [benchmark](https://github.com/ledisdb/ledisdb/wiki/Benchmark) for more.

## Todo

See [Issues todo](https://github.com/ledisdb/ledisdb/issues?labels=todo&page=1&state=open)

## Client

See [Clients](https://github.com/ledisdb/ledisdb/wiki/Clients) to find or contribute LedisDB client.

## Links

+ [Official Website](https://ledisdb.io)
+ [GoDoc](https://godoc.org/github.com/ledisdb/ledisdb)
+ [Server Commands](https://github.com/ledisdb/ledisdb/wiki/Commands)

## Caveat

+ Changing the backend database at runtime is very dangerous. Data validation is not guaranteed if this is done.

## Requirement

+ Go version >= 1.11

## Related Repos

+ [pika](https://github.com/Qihoo360/pika)


## Donate

If you like the project and want to buy me a cola, you can through: 

|PayPal|微信|
|------|---|
|[![](https://www.paypalobjects.com/webstatic/paypalme/images/pp_logo_small.png)](https://paypal.me/siddontang)|[![](https://github.com/siddontang/blog/blob/master/donate/weixin.png)|

## Feedback

+ Gmail: siddontang@gmail.com
