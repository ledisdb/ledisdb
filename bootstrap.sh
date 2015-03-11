#!/bin/bash

. ./dev.sh

# Test godep install
godep path > /dev/null 2>&1
if [ "$?" = 0 ]; then
    exit 0
fi

echo "Please use [godep](https://github.com/tools/godep) to build LedisDB, :-)"

go get -u github.com/szferi/gomdb
go get -u github.com/boltdb/bolt
go get -u github.com/ugorji/go/codec
go get -u github.com/BurntSushi/toml
go get -u github.com/edsrzf/mmap-go
go get -u github.com/syndtr/goleveldb/leveldb
go get -u github.com/cupcake/rdb

go get -u github.com/siddontang/go
go get -u github.com/siddontang/goredis
go get -u github.com/siddontang/rdb
