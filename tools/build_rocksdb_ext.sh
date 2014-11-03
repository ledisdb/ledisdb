#!/bin/bash

ROCKSDB_DIR=/usr/local/rocksdb

if test -z "$TARGET_OS"; then
    TARGET_OS=`uname -s`
fi

PLATFORM_SHARED_EXT="so"
PLATFORM_SHARED_LDFLAGS="-shared -Wl,-soname -Wl,"
PLATFORM_SHARED_CFLAGS="-fPIC"

if [ "$TARGET_OS" = "Darwin" ]; then
    PLATFORM_SHARED_EXT=dylib
    PLATFORM_SHARED_LDFLAGS="-dynamiclib -install_name "
fi

SONAME=librocksdb_ext.$PLATFORM_SHARED_EXT

g++ $PLATFORM_SHARED_LDFLAGS$SONAME $PLATFORM_SHARED_CFLAGS -std=c++0x -L$ROCKSDB_DIR/lib -lrocksdb -I/$ROCKSDB_DIR/include rocksdb_ext.cc -o $SONAME 