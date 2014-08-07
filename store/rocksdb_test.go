// +build rocksdb

package store

import (
	"github.com/siddontang/ledisdb/config"
	"os"

	"testing"
)

func newTestRocksDB() *DB {
	cfg := new(config.Config)
	cfg.DBName = "rocksdb"
	cfg.DataDir = "/tmp/testdb"

	os.RemoveAll(getStorePath(cfg))

	db, err := Open(cfg)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	return db
}

func TestRocksDB(t *testing.T) {
	db := newTestRocksDB()

	testStore(db, t)

	db.Close()
}
