#!/bin/sh

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

echo "GO_BUILD_TAGS=$GO_BUILD_TAGS" >> $OUTPUT