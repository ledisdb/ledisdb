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

	os.RemoveAll("/tmp/repl")
	os.MkdirAll("/tmp/repl", os.ModePerm)

	master, err = Open([]byte(`
        {
            "data_db" : {
                "path" : "/tmp/repl/master_db"
            },

            "binlog" : {
                "path" : "/tmp/repl/master_binlog"
            }   
        }
        `))
	if err != nil {
		t.Fatal(err)
	}

	slave, err = Open([]byte(`
        {
            "data_db" : {
                "path" : "/tmp/repl/slave_db"
            },

            "binlog" : {
                "path" : "/tmp/repl/slave_binlog"
            }   
        }
        `))
	if err != nil {
		t.Fatal(err)
	}

	db, _ := master.Select(0)
	db.Set([]byte("a"), []byte("1"))
	db.Set([]byte("b"), []byte("2"))
	db.Set([]byte("c"), []byte("3"))

	relayLog := "/tmp/repl/master_binlog/ledis-bin.0000001"

	var offset int64
	offset, err = slave.RepliateRelayLog(relayLog, 0)
	if err != nil {
		t.Fatal(err)
	} else {
		if st, err := os.Stat(relayLog); err != nil {
			t.Fatal(err)
		} else if st.Size() != offset {
			t.Fatal(st.Size(), offset)
		}
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
