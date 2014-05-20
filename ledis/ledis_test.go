package ledis

import (
	"sync"
	"testing"
)

var testLedis *Ledis
var testLedisOnce sync.Once

func getTestDB() *DB {
	f := func() {
		var d = []byte(`
            {
                "data_db" : {
                    "path" : "/tmp/testdb",
                    "compression":true,
                    "block_size" : 32768,
                    "write_buffer_size" : 2097152,
                    "cache_size" : 20971520
                }
            }
            `)
		var err error
		testLedis, err = Open(d)
		if err != nil {
			println(err.Error())
			panic(err)
		}

		testLedis.ldb.Clear()
	}

	testLedisOnce.Do(f)
	db, _ := testLedis.Select(0)
	return db
}

func TestDB(t *testing.T) {
	getTestDB()
}

func TestSelect(t *testing.T) {
	db0, _ := testLedis.Select(0)
	db1, _ := testLedis.Select(1)

	key0 := []byte("db0_test_key")
	key1 := []byte("db1_test_key")

	db0.Set(key0, []byte("0"))
	db1.Set(key1, []byte("1"))

	if v, err := db0.Get(key0); err != nil {
		t.Fatal(err)
	} else if string(v) != "0" {
		t.Fatal(string(v))
	}

	if v, err := db1.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "1" {
		t.Fatal(string(v))
	}
}
