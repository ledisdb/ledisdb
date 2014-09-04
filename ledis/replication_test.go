package ledis

import (
	"bytes"
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/store"
	"os"
	"path"
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
			return fmt.Errorf("replication error %d != %d", len(v), len(value))
		}
	}

	return nil
}

func TestReplication(t *testing.T) {
	var master *Ledis
	var slave *Ledis
	var err error

	cfgM := new(config.Config)
	cfgM.DataDir = "/tmp/test_repl/master"

	cfgM.BinLog.MaxFileNum = 10
	cfgM.BinLog.MaxFileSize = 50

	os.RemoveAll(cfgM.DataDir)

	master, err = Open(cfgM)
	if err != nil {
		t.Fatal(err)
	}

	cfgS := new(config.Config)
	cfgS.DataDir = "/tmp/test_repl/slave"

	os.RemoveAll(cfgS.DataDir)

	slave, err = Open(cfgS)
	if err != nil {
		t.Fatal(err)
	}

	db, _ := master.Select(0)
	db.Set([]byte("a"), []byte("value"))
	db.Set([]byte("b"), []byte("value"))
	db.Set([]byte("c"), []byte("value"))

	if tx, err := db.Begin(); err == nil {
		tx.HSet([]byte("a"), []byte("1"), []byte("value"))
		tx.HSet([]byte("b"), []byte("2"), []byte("value"))
		tx.HSet([]byte("c"), []byte("3"), []byte("value"))
		tx.Commit()
	} else {
		db.HSet([]byte("a"), []byte("1"), []byte("value"))
		db.HSet([]byte("b"), []byte("2"), []byte("value"))
		db.HSet([]byte("c"), []byte("3"), []byte("value"))
	}

	m, _ := db.Multi()
	m.Set([]byte("a1"), []byte("value"))
	m.Set([]byte("b1"), []byte("value"))
	m.Set([]byte("c1"), []byte("value"))
	m.Close()

	for _, name := range master.binlog.LogNames() {
		p := path.Join(master.binlog.LogPath(), name)

		err = slave.ReplicateFromBinLog(p)
		if err != nil {
			t.Fatal(err)
		}
	}

	if err = checkLedisEqual(master, slave); err != nil {
		t.Fatal(err)
	}

	slave.FlushAll()

	db.Set([]byte("a1"), []byte("value"))
	db.Set([]byte("b1"), []byte("value"))
	db.Set([]byte("c1"), []byte("value"))

	db.HSet([]byte("a1"), []byte("1"), []byte("value"))
	db.HSet([]byte("b1"), []byte("2"), []byte("value"))
	db.HSet([]byte("c1"), []byte("3"), []byte("value"))

	if tx, err := db.Begin(); err == nil {
		tx.HSet([]byte("a1"), []byte("1"), []byte("value1"))
		tx.HSet([]byte("b1"), []byte("2"), []byte("value1"))
		tx.HSet([]byte("c1"), []byte("3"), []byte("value1"))
		tx.Rollback()
	}

	info := new(MasterInfo)
	info.LogFileIndex = 1
	info.LogPos = 0
	var buf bytes.Buffer
	var n int

	for {
		buf.Reset()
		n, err = master.ReadEventsTo(info, &buf)
		if err != nil {
			t.Fatal(err)
		} else if info.LogFileIndex == -1 {
			t.Fatal("invalid log file index -1")
		} else if info.LogFileIndex == 0 {
			t.Fatal("invalid log file index 0")
		} else {
			if err = slave.ReplicateFromReader(&buf); err != nil {
				t.Fatal(err)
			}
			if n == 0 {
				break
			}
		}
	}

	if err = checkLedisEqual(master, slave); err != nil {
		t.Fatal(err)
	}
}
