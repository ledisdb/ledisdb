package wal

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestGoLevelDBStore(t *testing.T) {
	// Create a test dir
	dir, err := ioutil.TempDir("", "wal")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer os.RemoveAll(dir)

	// New level
	l, err := NewGoLevelDBStore(dir)
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer l.Close()

	testLogs(t, l)
}

func testLogs(t *testing.T, l Store) {
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
	if err := l.GetLog(10, &out); err.Error() != "log not found" {
		t.Fatalf("err: %v ", err)
	}

	// Write out a log
	log := Log{
		ID:   1,
		Data: []byte("first"),
	}
	for i := 1; i <= 10; i++ {
		log.ID = uint64(i)
		if err := l.StoreLog(&log); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Attempt to write multiple logs
	var logs []*Log
	for i := 11; i <= 20; i++ {
		nl := &Log{
			ID:   uint64(i),
			Data: []byte("first"),
		}
		logs = append(logs, nl)
	}
	if err := l.StoreLogs(logs); err != nil {
		t.Fatalf("err: %v", err)
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

	// Delete a suffix
	if err := l.Purge(5); err != nil {
		t.Fatalf("err: %v ", err)
	}

	// Verify they are all deleted
	for i := 1; i <= 5; i++ {
		if err := l.GetLog(uint64(i), &out); err != ErrLogNotFound {
			t.Fatalf("err: %v ", err)
		}
	}

	// Index should be one
	idx, err = l.FirstID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 6 {
		t.Fatalf("bad idx: %d", idx)
	}
	idx, err = l.LastID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 20 {
		t.Fatalf("bad idx: %d", idx)
	}

	// Should not be able to fetch
	if err := l.GetLog(5, &out); err != ErrLogNotFound {
		t.Fatalf("err: %v ", err)
	}

	if err := l.Clear(); err != nil {
		t.Fatal(err)
	}

	idx, err = l.FirstID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	idx, err = l.LastID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	now := uint32(time.Now().Unix())
	logs = []*Log{}
	for i := 1; i <= 20; i++ {
		nl := &Log{
			ID:         uint64(i),
			CreateTime: now - 20,
			Data:       []byte("first"),
		}
		logs = append(logs, nl)
	}

	if err := l.PurgeExpired(1); err != nil {
		t.Fatal(err)
	}

	idx, err = l.FirstID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}

	idx, err = l.LastID()
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	if idx != 0 {
		t.Fatalf("bad idx: %d", idx)
	}
}
