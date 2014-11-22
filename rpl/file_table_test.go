package rpl

import (
	"github.com/siddontang/go/log"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func TestFileTable(t *testing.T) {
	testFileTable(t, true)
	testFileTable(t, false)
}

func testFileTable(t *testing.T, useMmap bool) {
	log.SetLevel(log.LevelInfo)

	base, err := ioutil.TempDir("", "test_table")
	if err != nil {
		t.Fatal(err)
	}

	os.MkdirAll(base, 0755)

	defer os.RemoveAll(base)

	l := new(Log)
	l.Compression = 0
	l.Data = make([]byte, 4096)

	w := newTableWriter(base, 1, 1024*1024, useMmap)
	defer w.Close()

	for i := 0; i < 10; i++ {
		l.ID = uint64(i + 1)
		l.CreateTime = uint32(time.Now().Unix())

		l.Data[0] = byte(i + 1)

		if err := w.StoreLog(l); err != nil {
			t.Fatal(err)
		}
	}

	if w.first != 1 {
		t.Fatal(w.first)
	} else if w.last != 10 {
		t.Fatal(w.last)
	}

	l.ID = 10
	if err := w.StoreLog(l); err == nil {
		t.Fatal("must err")
	}

	var ll Log

	for i := 0; i < 10; i++ {
		if err := w.GetLog(uint64(i+1), &ll); err != nil {
			t.Fatal(err)
		} else if len(ll.Data) != 4096 {
			t.Fatal(len(ll.Data))
		} else if ll.Data[0] != byte(i+1) {
			t.Fatal(ll.Data[0])
		}
	}

	if err := w.GetLog(12, &ll); err == nil {
		t.Fatal("must nil")
	}

	var r *tableReader

	name := fmtTableDataName(w.base, w.index)

	if r, err = w.Flush(); err != nil {
		t.Fatal(err)
	}

	for i := 10; i < 20; i++ {
		l.ID = uint64(i + 1)
		l.CreateTime = uint32(time.Now().Unix())

		l.Data[0] = byte(i + 1)

		if err := w.StoreLog(l); err != nil {
			t.Fatal(err)
		}
	}

	if w.first != 11 {
		t.Fatal(w.first)
	} else if w.last != 20 {
		t.Fatal(w.last)
	}

	defer r.Close()

	for i := 0; i < 10; i++ {
		if err := r.GetLog(uint64(i+1), &ll); err != nil {
			t.Fatal(err)
		} else if len(ll.Data) != 4096 {
			t.Fatal(len(ll.Data))
		} else if ll.Data[0] != byte(i+1) {
			t.Fatal(ll.Data[0])
		}
	}

	if err := r.GetLog(12, &ll); err == nil {
		t.Fatal("must nil")
	}

	r.Close()

	if r, err = newTableReader(base, 1, useMmap); err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	for i := 0; i < 10; i++ {
		if err := r.GetLog(uint64(i+1), &ll); err != nil {
			t.Fatal(err)
		} else if len(ll.Data) != 4096 {
			t.Fatal(len(ll.Data))
		} else if ll.Data[0] != byte(i+1) {
			t.Fatal(ll.Data[0])
		}
	}

	if err := r.GetLog(12, &ll); err == nil {
		t.Fatal("must nil")
	}

	s := int64(r.data.Size())

	r.Close()

	log.SetLevel(log.LevelFatal)

	testRepair(t, name, 1, s, 11, useMmap)
	testRepair(t, name, 1, s, 20, useMmap)

	if err := os.Truncate(name, s-21); err != nil {
		t.Fatal(err)
	}

	if r, err := w.Flush(); err != nil {
		t.Fatal(err)
	} else {
		r.Close()
	}

	if r, err = newTableReader(base, 2, useMmap); err != nil {
		t.Fatal(err)
	}
	r.Close()
}

func testRepair(t *testing.T, name string, index int64, s int64, cutSize int64, useMmap bool) {
	var r *tableReader
	var err error

	if err := os.Truncate(name, s-cutSize); err != nil {
		t.Fatal(err)
	}

	if r, err = newTableReader(path.Dir(name), index, useMmap); err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	var ll Log
	for i := 0; i < 10; i++ {
		if err := r.GetLog(uint64(i+1), &ll); err != nil {
			t.Fatal(err, i)
		} else if len(ll.Data) != 4096 {
			t.Fatal(len(ll.Data))
		} else if ll.Data[0] != byte(i+1) {
			t.Fatal(ll.Data[0])
		}
	}

	if err := r.GetLog(12, &ll); err == nil {
		t.Fatal("must nil")
	}

	if s != int64(r.data.Size()) {
		t.Fatalf("repair error size %d != %d", s, r.data.Size())
	}

}
