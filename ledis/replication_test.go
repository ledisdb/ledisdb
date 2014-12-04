package ledis

import (
	"bytes"
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/store"
	"os"
	"testing"
)

func checkLedisEqual(master *Ledis, slave *Ledis) error {
	it := master.ldb.RangeLimitIterator(nil, nil, store.RangeClose, 0, -1)
	for ; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()

		if v, err := slave.ldb.Get(key); err != nil {
			return err
		} else if !bytes.Equal(v, value) {
			return fmt.Errorf("equal error at %q, %d != %d", key, len(v), len(value))
		}
	}

	return nil
}

func TestReplication(t *testing.T) {
	var master *Ledis
	var slave *Ledis
	var err error

	cfgM := config.NewConfigDefault()
	cfgM.DataDir = "/tmp/test_repl/master"

	cfgM.UseReplication = true
	cfgM.Replication.Compression = true

	os.RemoveAll(cfgM.DataDir)

	master, err = Open(cfgM)
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()

	cfgS := config.NewConfigDefault()
	cfgS.DataDir = "/tmp/test_repl/slave"
	cfgS.UseReplication = true
	cfgS.Readonly = true

	os.RemoveAll(cfgS.DataDir)

	slave, err = Open(cfgS)
	if err != nil {
		t.Fatal(err)
	}
	defer slave.Close()

	db, _ := master.Select(0)
	db.Set([]byte("a"), []byte("value"))
	db.Set([]byte("b"), []byte("value"))
	db.Set([]byte("c"), []byte("value"))

	db.HSet([]byte("a"), []byte("1"), []byte("value"))
	db.HSet([]byte("b"), []byte("2"), []byte("value"))
	db.HSet([]byte("c"), []byte("3"), []byte("value"))

	m, _ := db.Multi()
	m.Set([]byte("a1"), []byte("value"))
	m.Set([]byte("b1"), []byte("value"))
	m.Set([]byte("c1"), []byte("value"))
	m.Close()

	slave.FlushAll()

	db.Set([]byte("a1"), []byte("value"))
	db.Set([]byte("b1"), []byte("value"))
	db.Set([]byte("c1"), []byte("value"))

	db.HSet([]byte("a1"), []byte("1"), []byte("value"))
	db.HSet([]byte("b1"), []byte("2"), []byte("value"))
	db.HSet([]byte("c1"), []byte("3"), []byte("value"))

	var buf bytes.Buffer
	var n int
	var id uint64 = 1
	for {
		buf.Reset()
		n, id, err = master.ReadLogsTo(id, &buf)
		if err != nil {
			t.Fatal(err)
		} else if n != 0 {
			if err = slave.StoreLogsFromReader(&buf); err != nil {
				t.Fatal(err)
			}
		} else if n == 0 {
			break
		}
	}

	slave.WaitReplication()

	if err = checkLedisEqual(master, slave); err != nil {
		t.Fatal(err)
	}
}
