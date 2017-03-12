#!/bin/bash

OUTPUT=$1
PREFIX=$2
if test -z "$OUTPUT" || test -z "$PREFIX"; then
  echo "usage: $0 <output-filename> <directory_prefix>" >&2
  exit 1
fi

# Delete existing output, if it exists
rm -f $OUTPUT
touch $OUTPUT

source ./dev.sh

echo "CGO_CFLAGS=$CGO_CFLAGS" >> $OUTPUT
echo "CGO_CXXFLAGS=$CGO_CXXFLAGS" >> $OUTPUT
echo "CGO_LDFLAGS=$CGO_LDFLAGS" >> $OUTPUT
echo "LD_LIBRARY_PATH=$LD_LIBRARY_PATH" >> $OUTPUT
echo "DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH" >> $OUTPUT
echo "GO_BUILD_TAGS=$GO_BUILD_TAGS" >> $OUTPUT