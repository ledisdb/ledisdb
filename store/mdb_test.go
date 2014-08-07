package store

import (
	"github.com/siddontang/ledisdb/config"
	"os"

	"testing"
)

func newTestLMDB() *DB {
	cfg := new(config.Config)
	cfg.DBName = "lmdb"
	cfg.DataDir = "/tmp/testdb"
	cfg.LMDB.MapSize = 10 * 1024 * 1024

	os.RemoveAll(getStorePath(cfg))

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

func TestLMDBTx(t *testing.T) {
	db := newTestLMDB()

	testTx(db, t)

	db.Close()
}
