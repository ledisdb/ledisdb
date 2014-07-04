package ledis

import (
	"bytes"
	"github.com/siddontang/ledisdb/leveldb"
	"os"
	"testing"
)

func TestDump(t *testing.T) {
	os.RemoveAll("/tmp/test_ledis_master")
	os.RemoveAll("/tmp/test_ledis_slave")

	var masterConfig = []byte(`
    {
        "data_dir" : "/tmp/test_ledis_master",
        "data_db" : {
            "compression":true,
            "block_size" : 32768,
            "write_buffer_size" : 2097152,
            "cache_size" : 20971520
        }
    }
    `)

	master, err := OpenWithJsonConfig(masterConfig)
	if err != nil {
		t.Fatal(err)
	}

	var slaveConfig = []byte(`
    {
        "data_dir" : "/tmp/test_ledis_slave",
        "data_db" : {
            "compression":true,
            "block_size" : 32768,
            "write_buffer_size" : 2097152,
            "cache_size" : 20971520
        }
    }
    `)

	var slave *Ledis
	if slave, err = OpenWithJsonConfig(slaveConfig); err != nil {
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

	it := master.ldb.RangeLimitIterator(nil, nil, leveldb.RangeClose, 0, -1)
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
