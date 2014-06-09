package ledis

import (
	"bytes"
	"github.com/siddontang/go-leveldb/leveldb"
	"os"
	"testing"
)

func TestReplication(t *testing.T) {
	var master *Ledis
	var slave *Ledis
	var err error

	os.RemoveAll("/tmp/repl_repl")

	master, err = Open([]byte(`
        {
            "data_dir" : "/tmp/test_repl/master",
            "use_bin_log" : true
        }
        `))
	if err != nil {
		t.Fatal(err)
	}

	slave, err = Open([]byte(`
        {
            "data_dir" : "/tmp/test_repl/slave"
        }
        `))
	if err != nil {
		t.Fatal(err)
	}

	db, _ := master.Select(0)
	db.Set([]byte("a"), []byte("1"))
	db.Set([]byte("b"), []byte("2"))
	db.Set([]byte("c"), []byte("3"))

	binLogName := "/tmp/test_repl/master/bin_log/ledis-bin.0000001"

	err = slave.ReplicateFromBinLog(binLogName)
	if err != nil {
		t.Fatal(err)
	}

	it := master.ldb.Iterator(nil, nil, leveldb.RangeClose, 0, -1)
	for ; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()

		if v, err := slave.ldb.Get(key); err != nil {
			t.Fatal(err)
		} else if !bytes.Equal(v, value) {
			t.Fatal("replication error", len(v), len(value))
		}
	}
}
