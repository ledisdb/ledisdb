// +build !windows

package store

import (
	"github.com/siddontang/ledisdb/store/boltdb"
)

func init() {
	Register(boltdb.Store{})
}
