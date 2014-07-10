# ledisdb

Ledisdb is a high performance NoSQL like Redis based on LevelDB written by go. It supports some advanced data structure like kv, list, hash and zset, and may be alternative for Redis.

## Build and Install

+ Create a workspace and checkout ledisdb source

        mkdir $WORKSPACE
        cd $WORKSPACE
        git clone git@github.com:siddontang/ledisdb.git src/github.com/siddontang/ledisdb

        cd src/github.com/siddontang/ledisdb

+ Install leveldb and snappy, if you have installed, skip.

    I supply a simple shell to install leveldb and snappy, you can use: 

        sh build_leveldb.sh

    It will default install leveldb at /usr/local/leveldb and snappy at /usr/local/snappy

+ Change LEVELDB_DIR and SNAPPY_DIR to real install path in dev.sh.

+ Then:

        . ./bootstap.sh 
        . ./dev.sh

        go install ./...

## Run

    ./ledis-server -config=/etc/ledis.json

    //another shell
    ledis-cli -p 6380
    
    ledis 127.0.0.1:6380> set a 1
    OK
    ledis 127.0.0.1:6380> get a
    "1"

## Lib
    
    import "github.com/siddontang/ledisdb/ledis"
    l, _ := ledis.Open(cfg)
    db, _ := l.Select(0)

    db.Set(key, value)

    db.Get(key)


## Replication

set slaveof in config or dynamiclly

    ledis-cli -p 6381 

    ledis 127.0.0.1:6381> slaveof 127.0.0.1:6380
    OK

## Benchmark

See benchmark.md for more.

## Todo

+ Admin

## GoDoc

[![GoDoc](https://godoc.org/github.com/siddontang/ledisdb?status.png)](https://godoc.org/github.com/siddontang/ledisdb)

## Commands

Some server commands explaintions are [here](https://github.com/siddontang/ledisdb/wiki/Commands), others will add continuate.

## Thanks

Gmail: cenqichao@gmail.com

Gmail: chendahui007@gmail.com

## Feedback

Gmail: siddontang@gmail.com