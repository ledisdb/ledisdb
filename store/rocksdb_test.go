// +build rocksdb

package store

import (
	"os"
	"testing"
)

func newTestRocksDB() *DB {
	cfg := new(Config)
	cfg.Name = RocksDBName
	cfg.Path = "/tmp/testdb/rocksdb"

	os.RemoveAll(cfg.Path)

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
