package ledis

import (
	"sync"
	"testing"
)

var testDB *DB
var testDBOnce sync.Once

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
		db, err := OpenDB(d)
		if err != nil {
			println(err.Error())
			panic(err)
		}

		testDB = db

		testDB.db.Clear()
	}

	testDBOnce.Do(f)
	return testDB
}

func TestDB(t *testing.T) {
	getTestDB()
}
