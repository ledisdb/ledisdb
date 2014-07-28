// +build leveldb

package store

import (
	"os"
	"testing"
)

func newTestLevelDB() *DB {
	cfg := new(Config)
	cfg.Name = LevelDBName
	cfg.Path = "/tmp/testdb/leveldb"

	os.RemoveAll(cfg.Path)

	db, err := Open(cfg)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	return db
}

func TestLevelDB(t *testing.T) {
	db := newTestLevelDB()

	testStore(db, t)

	db.Close()
}
