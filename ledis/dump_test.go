package ledis

import (
	"bytes"
	"github.com/siddontang/go-leveldb/leveldb"
	"os"
	"testing"
)

func TestDump(t *testing.T) {
	os.RemoveAll("/tmp/testdb_master")
	os.RemoveAll("/tmp/testdb_slave")
	os.Remove("/tmp/testdb.dump")

	var masterConfig = []byte(`
    {
        "data_db" : {
            "path" : "/tmp/testdb_master",
            "compression":true,
            "block_size" : 32768,
            "write_buffer_size" : 2097152,
            "cache_size" : 20971520
        }
    }
    `)

	master, err := Open(masterConfig)
	if err != nil {
		t.Fatal(err)
	}

	var slaveConfig = []byte(`
    {
        "data_db" : {
            "path" : "/tmp/testdb_slave",
            "compression":true,
            "block_size" : 32768,
            "write_buffer_size" : 2097152,
            "cache_size" : 20971520
        }
    }
    `)

	var slave *Ledis
	if slave, err = Open(slaveConfig); err != nil {
		t.Fatal(err)
	}

	db, _ := master.Select(0)

	db.Set([]byte("a"), []byte("1"))
	db.Set([]byte("b"), []byte("2"))
	db.Set([]byte("c"), []byte("3"))

	if err := master.DumpFile("/tmp/testdb.dump"); err != nil {
		t.Fatal(err)
	}

	if _, err := slave.LoadDumpFile("/tmp/testdb.dump"); err != nil {
		t.Fatal(err)
	}

	it := master.ldb.Iterator(nil, nil, leveldb.RangeClose, 0, -1)
	for ; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()

		if v, err := slave.ldb.Get(key); err != nil {
			t.Fatal(err)
		} else if !bytes.Equal(v, value) {
			t.Fatal("load dump error")
		}
	}
}
