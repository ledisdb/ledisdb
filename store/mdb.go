package store

import (
	"github.com/siddontang/ledisdb/store/mdb"
)

func init() {
	Register(mdb.Store{})
}
