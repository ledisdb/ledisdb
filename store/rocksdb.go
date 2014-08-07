// +build rocksdb

package store

import (
	"github.com/siddontang/ledisdb/store/rocksdb"
)

func init() {
	Register(rocksdb.Store{})
}
