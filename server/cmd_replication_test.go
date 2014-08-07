package server

import (
	"bytes"
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/store"
	"os"
	"testing"
	"time"
)

func checkDataEqual(master *App, slave *App) error {
	it := master.ldb.DataDB().RangeLimitIterator(nil, nil, store.RangeClose, 0, -1)
	for ; it.Valid(); it.Next() {
		key := it.Key()
		value := it.Value()

		if v, err := slave.ldb.DataDB().Get(key); err != nil {
			return err
		} else if !bytes.Equal(v, value) {
			return fmt.Errorf("replication error %d != %d", len(v), len(value))
		}
	}

	return nil
}

func TestReplication(t *testing.T) {
	data_dir := "/tmp/test_replication"
	os.RemoveAll(data_dir)

	masterCfg := new(config.Config)
	masterCfg.DataDir = fmt.Sprintf("%s/master", data_dir)
	masterCfg.Addr = "127.0.0.1:11182"
	masterCfg.BinLog.MaxFileSize = 1 * 1024 * 1024
	masterCfg.BinLog.MaxFileNum = 10

	var master *App
	var slave *App
	var err error
	master, err = NewApp(masterCfg)
	if err != nil {
		t.Fatal(err)
	}

	slaveCfg := new(config.Config)
	slaveCfg.DataDir = fmt.Sprintf("%s/slave", data_dir)
	slaveCfg.Addr = "127.0.0.1:11183"
	slaveCfg.SlaveOf = masterCfg.Addr

	slave, err = NewApp(slaveCfg)
	if err != nil {
		t.Fatal(err)
	}

	go master.Run()

	db, _ := master.ldb.Select(0)

	value := make([]byte, 10)

	db.Set([]byte("a"), value)
	db.Set([]byte("b"), value)
	db.HSet([]byte("a"), []byte("1"), value)
	db.HSet([]byte("b"), []byte("2"), value)

	go slave.Run()

	time.Sleep(1 * time.Second)

	if err = checkDataEqual(master, slave); err != nil {
		t.Fatal(err)
	}

	db.Set([]byte("a1"), value)
	db.Set([]byte("b1"), value)
	db.HSet([]byte("a1"), []byte("1"), value)
	db.HSet([]byte("b1"), []byte("2"), value)

	time.Sleep(1 * time.Second)
	if err = checkDataEqual(master, slave); err != nil {
		t.Fatal(err)
	}

	slave.slaveof("")

	db.Set([]byte("a2"), value)
	db.Set([]byte("b2"), value)
	db.HSet([]byte("a2"), []byte("1"), value)
	db.HSet([]byte("b2"), []byte("2"), value)

	db.Set([]byte("a3"), value)
	db.Set([]byte("b3"), value)
	db.HSet([]byte("a3"), []byte("1"), value)
	db.HSet([]byte("b3"), []byte("2"), value)

	if err = checkDataEqual(master, slave); err == nil {
		t.Fatal("must error")
	}

	slave.slaveof(masterCfg.Addr)
	time.Sleep(1 * time.Second)

	if err = checkDataEqual(master, slave); err != nil {
		t.Fatal(err)
	}

}
