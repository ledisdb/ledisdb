package store

import (
	"github.com/siddontang/ledisdb/config"
	"os"
	"testing"
)

func newTestBoltDB() *DB {
	cfg := new(config.Config)
	cfg.DBName = "boltdb"
	cfg.DataDir = "/tmp/testdb"

	os.RemoveAll(getStorePath(cfg))

	db, err := Open(cfg)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	return db
}

func TestBoltDB(t *testing.T) {
	db := newTestBoltDB()

	testStore(db, t)

	db.Close()
}

func TestBoltDBTx(t *testing.T) {
	db := newTestBoltDB()

	testTx(db, t)

	db.Close()
}
