package store

import (
	"github.com/siddontang/ledisdb/config"
	"os"

	"testing"
)

func newTestGoLevelDB() *DB {
	cfg := new(config.Config)
	cfg.DBName = "goleveldb"
	cfg.DataDir = "/tmp/testdb"

	os.RemoveAll(getStorePath(cfg))

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
