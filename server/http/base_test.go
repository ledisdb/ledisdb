package http

import (
	"github.com/siddontang/ledisdb/ledis"
	"os"
	"sync"
)

var once sync.Once
var ldb *ledis.Ledis

func getTestDB() *ledis.DB {
	f := func() {
		var err error
		if _, err = os.Stat("/tmp/test_http_api_db"); err == nil {
			if err = os.RemoveAll("/tmp/test_http_api_db"); err != nil {
				panic(err)
			}
		} else if !os.IsNotExist(err) {
			panic(err)
		}
		var cfg ledis.Config
		cfg.DataDir = "/tmp/test_http_api_db"
		cfg.DataDB.BlockSize = 32768
		cfg.DataDB.WriteBufferSize = 20971520
		cfg.DataDB.CacheSize = 20971520
		cfg.DataDB.Compression = true

		ldb, err = ledis.Open(&cfg)
		if err != nil {
			panic(err)
		}
	}
	once.Do(f)
	db, err := ldb.Select(0)
	if err != nil {
		panic(err)
	}
	return db
}
