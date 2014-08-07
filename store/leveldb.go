// +build leveldb

package store

import (
	"github.com/siddontang/ledisdb/store/leveldb"
)

func init() {
	Register(leveldb.Store{})
}
