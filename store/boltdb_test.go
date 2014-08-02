package store

import (
	"os"
	"testing"
)

func newTestBoltDB() *DB {
	cfg := new(Config)
	cfg.Name = BoltDBName
	cfg.Path = "/tmp/testdb/boltdb"

	os.RemoveAll(cfg.Path)

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
