package rpl

import (
	"github.com/siddontang/ledisdb/config"
	"io/ioutil"
	"os"
	"testing"
)

func TestGoLevelDBStore(t *testing.T) {
	// Create a test dir
	dir, err := ioutil.TempDir("", "wal")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer os.RemoveAll(dir)

	// New level
	l, err := NewGoLevelDBStore(dir, 0)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer l.Close()

	testLogs(t, l)
}

func TestFileStore(t *testing.T) {
	// Create a test dir
	dir, err := ioutil.TempDir("", "ldb")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer os.RemoveAll(dir)

	// New level
	cfg := config.NewConfigDefault()
	cfg.Replication.MaxLogFileSize = 4096

	l, err := NewFileStore(dir, cfg)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer l.Close()

	testLogs(t, l)
}

func testLogs(t *testing.T, l LogStore) {
	// Should be no first index
	idx, err := l.FirstID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}
	// Should be no last index
	idx, err = l.LastID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Try a filed fetch
	var out Log
	if err := l.GetLog(10, &out); err != ErrLogNotFound {
		t.Fatalf("err: %v ", err)
	}

	data := make([]byte, 1024)

	// Write out a log
	log := Log{
		ID:   1,
		Data: data,
	}
	for i := 1; i <= 10; i++ {
		log.ID = uint64(i)
		if err := l.StoreLog(&log); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Attempt to write multiple logs
	for i := 11; i <= 20; i++ {
		nl := &Log{
			ID:   uint64(i),
			Data: data,
		}

		if err := l.StoreLog(nl); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Try to fetch
	if err := l.GetLog(1, &out); err != nil {
		t.Fatalf("err: %v ", err)
	}

	// Try to fetch
	if err := l.GetLog(10, &out); err != nil {
		t.Fatalf("err: %v ", err)
	}

	// Try to fetch
	if err := l.GetLog(20, &out); err != nil {
		t.Fatalf("err: %v ", err)
	}

	// Check the lowest index
	idx, err = l.FirstID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 1 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Check the highest index
	idx, err = l.LastID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 20 {
		t.Fatalf("bad idx: %d", idx)
	}

	if err = l.Clear(); err != nil {
		t.Fatalf("err :%v", err)
	}

	// Check the lowest index
	idx, err = l.FirstID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Check the highest index
	idx, err = l.LastID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Write out a log
	log = Log{
		ID:   1,
		Data: data,
	}
	for i := 1; i <= 10; i++ {
		log.ID = uint64(i)
		if err := l.StoreLog(&log); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

}
