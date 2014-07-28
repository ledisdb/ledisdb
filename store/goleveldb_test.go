package store

import (
	"os"
	"testing"
)

func newTestGoLevelDB() *DB {
	cfg := new(Config)
	cfg.Name = GoLevelDBName
	cfg.Path = "/tmp/testdb/goleveldb"

	os.RemoveAll(cfg.Path)

	db, err := Open(cfg)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	return db
}

func TestGoLevelDB(t *testing.T) {
	db := newTestGoLevelDB()

	testStore(db, t)

	db.Close()
}
