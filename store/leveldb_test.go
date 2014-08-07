// +build leveldb

package store

import (
	"github.com/siddontang/ledisdb/config"

	"os"
	"testing"
)

func newTestLevelDB() *DB {
	cfg := new(config.Config)
	cfg.DBName = "leveldb"
	cfg.DataDir = "/tmp/testdb"

	os.RemoveAll(getStorePath(cfg))

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
