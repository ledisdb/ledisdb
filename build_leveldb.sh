#!/bin/bash

SNAPPY_DIR=/usr/local/snappy
LEVELDB_DIR=/usr/local/leveldb

mkdir -p ./build

cd ./build

if [ ! -f $SNAPPY_DIR/lib/libsnappy.a ]; then
    (git clone git@github.com:siddontang/snappy.git && \
        cd ./snappy && \
        ./configure --prefix=$SNAPPY_DIR && \
        make && \
        make install)
else
    echo "skip install snappy"
fi

if [ ! -f $LEVELDB_DIR/lib/libleveldb.a ]; then
    (git clone git@github.com:siddontang/leveldb.git && \
        cd ./leveldb && \
        echo "echo \"PLATFORM_CFLAGS+=-I$SNAPPY_DIR/include\" >> build_config.mk" >> build_detect_platform &&
        echo "echo \"PLATFORM_CXXFLAGS+=-I$SNAPPY_DIR/include\" >> build_config.mk" >> build_detect_platform &&
        echo "echo \"PLATFORM_LDFLAGS+=-L $SNAPPY_DIR/lib -lsnappy\" >> build_config.mk" >> build_detect_platform &&
        make SNAPPY=1 && \
        make && \
        mkdir -p $LEVELDB_DIR/include/leveldb && \
        install include/leveldb/*.h $LEVELDB_DIR/include/leveldb && \
        mkdir -p $LEVELDB_DIR/lib && \
        cp -f libleveldb.* $LEVELDB_DIR/lib)
else
    echo "skip install leveldb"
fi