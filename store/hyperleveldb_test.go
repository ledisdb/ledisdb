// +build hyperleveldb

package store

import (
	"github.com/siddontang/ledisdb/config"
	"os"
	"testing"
)

func newTestHyperLevelDB() *DB {
	cfg := new(config.Config)
	cfg.DBName = "hyperleveldb"
	cfg.DataDir = "/tmp/testdb"

	os.RemoveAll(getStorePath(cfg))

	db, err := Open(cfg)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	return db
}

func TestHyperLevelDB(t *testing.T) {
	db := newTestHyperLevelDB()

	testStore(db, t)

	db.Close()
}
