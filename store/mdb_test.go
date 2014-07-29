// +build !windows

package store

import (
	"os"
	"testing"
)

func newTestLMDB() *DB {
	cfg := new(Config)
	cfg.Name = LMDBName
	cfg.Path = "/tmp/testdb/lmdb"

	cfg.MapSize = 20 * 1024 * 1024

	os.RemoveAll(cfg.Path)

	db, err := Open(cfg)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	return db
}

func TestLMDB(t *testing.T) {
	db := newTestLMDB()

	testStore(db, t)

	db.Close()
}
