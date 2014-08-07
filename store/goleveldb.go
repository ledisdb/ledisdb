package store

import (
	"github.com/siddontang/ledisdb/store/goleveldb"
)

func init() {
	Register(goleveldb.Store{})
}
