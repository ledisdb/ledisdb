#!/bin/bash

. ./dev.sh

# Test godep install
godep path > /dev/null 2>&1
if [ "$?" = 0 ]; then
    exit 0
fi

echo "Please use [godep](https://github.com/tools/godep) to build LedisDB, :-)"

go get -d ./...
