#!/bin/bash

. ./dev.sh

# Test godep install
godep path > /dev/null 2>&1
if [ "$?" = 0 ]; then
    GOPATH=`godep path`
    # https://github.com/tools/godep/issues/60
    # have to rm Godeps/_workspace first, then restore
    rm -rf $GOPATH
    godep restore
    exit 0
fi

go get -u github.com/szferi/gomdb

go get -u github.com/boltdb/bolt

go get -u github.com/ugorji/go/codec
go get -u github.com/BurntSushi/toml
go get -u github.com/edsrzf/mmap-go
go get -u github.com/syndtr/goleveldb/leveldb

go get -u github.com/siddontang/go/bson
go get -u github.com/siddontang/go/log
go get -u github.com/siddontang/go/snappy
go get -u github.com/siddontang/go/num
go get -u github.com/siddontang/go/filelock
go get -u github.com/siddontang/go/sync2
go get -u github.com/siddontang/go/arena
