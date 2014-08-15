// +build hyperleveldb

package store

import (
	"github.com/siddontang/ledisdb/store/hyperleveldb"
)

func init() {
	Register(hyperleveldb.Store{})
}
