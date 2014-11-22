package store

import (
	"bytes"
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/store/driver"
	"os"
	"reflect"
	"testing"
)

func TestStore(t *testing.T) {
	cfg := config.NewConfigDefault()
	cfg.DataDir = "/tmp/testdb"
	cfg.LMDB.MapSize = 10 * 1024 * 1024

	ns := driver.ListStores()
	for _, s := range ns {
		cfg.DBName = s

		os.RemoveAll(getStorePath(cfg))

		db, err := Open(cfg)
		if err != nil {
			t.Fatal(err)
		}

		testStore(db, t)
		testClear(db, t)
		testTx(db, t)

		db.Close()
	}
}

func testStore(db *DB, t *testing.T) {
	testSimple(db, t)
	testBatch(db, t)
	testIterator(db, t)
	testSnapshot(db, t)
	testBatchData(db, t)
}

func testClear(db *DB, t *testing.T) {
	it := db.RangeIterator(nil, nil, RangeClose)
	for ; it.Valid(); it.Next() {
		db.Delete(it.RawKey())
	}
	it.Close()
}

func testSimple(db *DB, t *testing.T) {
	key := []byte("key")
	value := []byte("hello world")
	if err := db.Put(key, value); err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get(key); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(v, value) {
		t.Fatal("not equal")
	}

	if v, err := db.GetSlice(key); err != nil {
		t.Fatal(err)
	} else if v == nil {
		t.Fatal("must not nil")
	} else if !bytes.Equal(v.Data(), value) {
		t.Fatal("not equal")
	} else {
		v.Free()
	}

	if err := db.Delete(key); err != nil {
		t.Fatal(err)
	}
	if v, err := db.Get(key); err != nil {
		t.Fatal(err)
	} else if v != nil {
		t.Fatal("must nil")
	}

	if v, err := db.GetSlice(key); err != nil {
		t.Fatal(err)
	} else if v != nil {
		t.Fatal("must nil")
	}

	if err := db.Put(key, nil); err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get(key); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(v, []byte{}) {
		t.Fatal("must empty")
	}
}

func testBatch(db *DB, t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")

	value := []byte("hello world")

	db.Put(key1, value)
	db.Put(key2, value)

	wb := db.NewWriteBatch()

	wb.Delete(key2)
	wb.Put(key1, []byte("hello world2"))

	if err := wb.Commit(); err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get(key2); err != nil {
		t.Fatal(err)
	} else if v != nil {
		t.Fatal("must nil")
	}

	if v, err := db.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "hello world2" {
		t.Fatal(string(v))
	}

	wb.Delete(key1)

	wb.Rollback()

	if v, err := db.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "hello world2" {
		t.Fatal(string(v))
	}

	wb.Put(key1, nil)
	wb.Put(key2, []byte{})

	if err := wb.Commit(); err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get(key1); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(v, []byte{}) {
		t.Fatal("must empty")
	}

	if v, err := db.Get(key2); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(v, []byte{}) {
		t.Fatal("must empty")
	}

	db.Delete(key1)
	db.Delete(key2)
}

func checkIterator(it *RangeLimitIterator, cv ...int) error {
	v := make([]string, 0, len(cv))
	for ; it.Valid(); it.Next() {
		k := it.Key()
		if string(it.Value()) != "value" {
			return fmt.Errorf("invalid value")
		}
		v = append(v, string(k))
	}

	it.Close()

	if len(v) != len(cv) {
		return fmt.Errorf("len error %d != %d", len(v), len(cv))
	}

	for k, i := range cv {
		if fmt.Sprintf("key_%d", i) != v[k] {
			return fmt.Errorf("%s, %d", v[k], i)
		}
	}

	return nil
}

func testIterator(db *DB, t *testing.T) {
	i := db.NewIterator()
	for i.SeekToFirst(); i.Valid(); i.Next() {
		db.Delete(i.Key())
	}
	i.Close()

	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		value := []byte("value")
		db.Put(key, value)
	}

	i = db.NewIterator()
	i.SeekToFirst()

	if !i.Valid() {
		t.Fatal("must valid")
	} else if string(i.Key()) != "key_0" {
		t.Fatal(string(i.Key()))
	}
	i.Close()

	var it *RangeLimitIterator

	k := func(i int) []byte {
		return []byte(fmt.Sprintf("key_%d", i))
	}

	it = db.RangeLimitIterator(k(1), k(5), RangeClose, 0, -1)
	if err := checkIterator(it, 1, 2, 3, 4, 5); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RangeLimitIterator(k(1), k(5), RangeClose, 0, -1)
	if err := checkIterator(it, 1, 2, 3, 4, 5); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RangeLimitIterator(k(1), k(5), RangeClose, 1, 3)
	if err := checkIterator(it, 2, 3, 4); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RangeLimitIterator(k(1), k(5), RangeLOpen, 0, -1)
	if err := checkIterator(it, 2, 3, 4, 5); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RangeLimitIterator(k(1), k(5), RangeROpen, 0, -1)
	if err := checkIterator(it, 1, 2, 3, 4); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RangeLimitIterator(k(1), k(5), RangeOpen, 0, -1)
	if err := checkIterator(it, 2, 3, 4); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RevRangeLimitIterator(k(1), k(5), RangeClose, 0, -1)
	if err := checkIterator(it, 5, 4, 3, 2, 1); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RevRangeLimitIterator(k(1), k(5), RangeClose, 1, 3)
	if err := checkIterator(it, 4, 3, 2); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RevRangeLimitIterator(k(1), k(5), RangeLOpen, 0, -1)
	if err := checkIterator(it, 5, 4, 3, 2); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RevRangeLimitIterator(k(1), k(5), RangeROpen, 0, -1)
	if err := checkIterator(it, 4, 3, 2, 1); err != nil {
		t.Fatal(err)
	}
	it.Close()

	it = db.RevRangeLimitIterator(k(1), k(5), RangeOpen, 0, -1)
	if err := checkIterator(it, 4, 3, 2); err != nil {
		t.Fatal(err)
	}
	it.Close()
}

func testSnapshot(db *DB, t *testing.T) {
	foo := []byte("foo")
	bar := []byte("bar")
	v1 := []byte("v1")
	v2 := []byte("v2")

	db.Put(foo, v1)
	db.Put(bar, v1)

	snap, err := db.NewSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	i := snap.NewIterator()

	i.Seek([]byte("foo"))

	if !i.Valid() {
		t.Fatal("must valid")
	} else if string(i.Value()) != "v1" {
		t.Fatal(string(i.Value()))
	}
	i.Close()

	db.Put(foo, v2)
	db.Put(bar, v2)

	if v, err := snap.Get(foo); err != nil {
		t.Fatal(err)
	} else if string(v) != "v1" {
		t.Fatal(string(v))
	}

	if v, err := snap.Get(bar); err != nil {
		t.Fatal(err)
	} else if string(v) != "v1" {
		t.Fatal(string(v))
	}

	if v, err := db.Get(foo); err != nil {
		t.Fatal(err)
	} else if string(v) != "v2" {
		t.Fatal(string(v))
	}

	if v, err := db.Get(bar); err != nil {
		t.Fatal(err)
	} else if string(v) != "v2" {
		t.Fatal(string(v))
	}

	snap.Close()

	if v, err := db.Get(foo); err != nil {
		t.Fatal(err)
	} else if string(v) != "v2" {
		t.Fatal(string(v))
	}

}

func testBatchData(db *DB, t *testing.T) {
	w := db.NewWriteBatch()

	w.Put([]byte("a"), []byte("1"))
	w.Put([]byte("b"), nil)
	w.Delete([]byte("c"))

	d := w.BatchData()

	if kvs, err := d.Items(); err != nil {
		t.Fatal(err)
	} else if len(kvs) != 3 {
		t.Fatal(len(kvs))
	} else if !reflect.DeepEqual(kvs[0], BatchItem{[]byte("a"), []byte("1")}) {
		t.Fatal("must equal")
	} else if !reflect.DeepEqual(kvs[1], BatchItem{[]byte("b"), []byte{}}) {
		t.Fatal("must equal")
	} else if !reflect.DeepEqual(kvs[2], BatchItem{[]byte("c"), nil}) {
		t.Fatal("must equal")
	}

	if err := d.Append(d); err != nil {
		t.Fatal(err)
	} else if d.Len() != 6 {
		t.Fatal(d.Len())
	}

	if kvs, err := d.Items(); err != nil {
		t.Fatal(err)
	} else if len(kvs) != 6 {
		t.Fatal(len(kvs))
	} else if !reflect.DeepEqual(kvs[0], BatchItem{[]byte("a"), []byte("1")}) {
		t.Fatal("must equal")
	} else if !reflect.DeepEqual(kvs[1], BatchItem{[]byte("b"), []byte{}}) {
		t.Fatal("must equal")
	} else if !reflect.DeepEqual(kvs[2], BatchItem{[]byte("c"), nil}) {
		t.Fatal("must equal")
	} else if !reflect.DeepEqual(kvs[3], BatchItem{[]byte("a"), []byte("1")}) {
		t.Fatal("must equal")
	} else if !reflect.DeepEqual(kvs[4], BatchItem{[]byte("b"), []byte{}}) {
		t.Fatal("must equal")
	} else if !reflect.DeepEqual(kvs[5], BatchItem{[]byte("c"), nil}) {
		t.Fatal("must equal")
	}
}
