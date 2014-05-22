# ledisdb

Ledisdb is a high performance nosql like redis based on leveldb written by go. It's supports some advanced data structure like kv, list, hash and zset.

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

## Benchmark

See benchmark.md for more.

## Todo

+ Expire + TTL
+ Replication
+ Admin

## Feedback

Gmail: siddontang@gmail.com