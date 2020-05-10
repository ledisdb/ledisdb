#!/bin/bash

export LEDISTOP=$(pwd)
export LEDISROOT="${LEDISROOT:-${LEDISTOP/\/src\/github.com\/ledisdb\/ledisdb/}}"
# LEDISTOP sanity check
if [[ "$LEDISTOP" == "${LEDISTOP/\/src\/github.com\/ledisdb\/ledisdb/}" ]]; then
    echo "WARNING: LEDISTOP($LEDISTOP) does not contain src/github.com/ledisdb/ledisdb"
    false
fi

#default snappy and leveldb install path
#you may change yourself
HOMEBREW_PREFIX=$(type brew >/dev/null 2>&1 && brew --prefix 2>/dev/null)
if [[ ! -z "$HOMEBREW_PREFIX" ]]; then
  SNAPPY_DIR=$HOMEBREW_PREFIX/opt/snappy
  LEVELDB_DIR=$HOMEBREW_PREFIX/opt/leveldb
  ROCKSDB_DIR=$HOMEBREW_PREFIX/opt/rocksdb
else
  SNAPPY_DIR=/usr/local/snappy
  LEVELDB_DIR=/usr/local/leveldb
  ROCKSDB_DIR=/usr/local/rocksdb
fi

function add_path()
{
  # $1 path variable
  # $2 path to add
  if [ -d "$2" ] && [[ ":$1:" != *":$2:"* ]]; then
    echo "$1:$2"
  else
    echo "$1"
  fi
}

export GOPATH=$(add_path $GOPATH $LEDISROOT)

GO_BUILD_TAGS=
CGO_CFLAGS=
CGO_CXXFLAGS=
CGO_LDFLAGS=

# check dependent libray, now we only check simply, maybe later add proper checking way.

# check snappy 
if [ -f $SNAPPY_DIR/include/snappy.h ]; then
    CGO_CFLAGS="$CGO_CFLAGS -I$SNAPPY_DIR/include"
    CGO_CXXFLAGS="$CGO_CXXFLAGS -I$SNAPPY_DIR/include"
    CGO_LDFLAGS="$CGO_LDFLAGS -L$SNAPPY_DIR/lib -lsnappy"
    LD_LIBRARY_PATH=$(add_path $LD_LIBRARY_PATH $SNAPPY_DIR/lib)
    DYLD_LIBRARY_PATH=$(add_path $DYLD_LIBRARY_PATH $SNAPPY_DIR/lib)
fi

# check leveldb
if [ -f $LEVELDB_DIR/include/leveldb/c.h ]; then
    CGO_CFLAGS="$CGO_CFLAGS -I$LEVELDB_DIR/include"
    CGO_CXXFLAGS="$CGO_CXXFLAGS -I$LEVELDB_DIR/include"
    CGO_LDFLAGS="$CGO_LDFLAGS -L$LEVELDB_DIR/lib -lleveldb"
    LD_LIBRARY_PATH=$(add_path $LD_LIBRARY_PATH $LEVELDB_DIR/lib)
    DYLD_LIBRARY_PATH=$(add_path $DYLD_LIBRARY_PATH $LEVELDB_DIR/lib)
    GO_BUILD_TAGS="$GO_BUILD_TAGS leveldb"
fi

# check rocksdb
if [ -f $ROCKSDB_DIR/include/rocksdb/c.h ]; then
    CGO_CFLAGS="$CGO_CFLAGS -I$ROCKSDB_DIR/include"
    CGO_CXXFLAGS="$CGO_CXXFLAGS -I$ROCKSDB_DIR/include"
    CGO_LDFLAGS="$CGO_LDFLAGS -L$ROCKSDB_DIR/lib -lrocksdb"
    LD_LIBRARY_PATH=$(add_path $LD_LIBRARY_PATH $ROCKSDB_DIR/lib)
    DYLD_LIBRARY_PATH=$(add_path $DYLD_LIBRARY_PATH $ROCKSDB_DIR/lib)
    GO_BUILD_TAGS="$GO_BUILD_TAGS rocksdb"
fi

export CGO_CFLAGS
export CGO_CXXFLAGS
export CGO_LDFLAGS
export LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH
export GO_BUILD_TAGS
