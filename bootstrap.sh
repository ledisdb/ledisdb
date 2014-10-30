#!/bin/bash

. ./dev.sh

# Test godep install
godep path > /dev/null 2>&1
if [ "$?" = 0 ]; then
    GOPATH=`godep path`
    godep restore
    exit 0
fi

go get github.com/siddontang/goleveldb/leveldb

go get github.com/szferi/gomdb

go get github.com/boltdb/bolt

go get github.com/ugorji/go/codec
go get github.com/BurntSushi/toml


go get github.com/siddontang/go/bson
go get github.com/siddontang/go/log
go get github.com/siddontang/go/snappy
go get github.com/siddontang/go/num
go get github.com/siddontang/go/filelock
go get github.com/siddontang/go/sync2
go get github.com/siddontang/go/arena