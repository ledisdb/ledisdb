package store

import (
	"github.com/siddontang/ledisdb/store/driver"
	"testing"
)

func TestTx(t *testing.T) {

}

func testTx(db *DB, t *testing.T) {
	if tx, err := db.Begin(); err != nil {
		if err == driver.ErrTxSupport {
			return
		} else {
			t.Fatal(err)
		}
	} else {
		tx.Rollback()
	}

	key1 := []byte("1")
	key2 := []byte("2")
	key3 := []byte("3")
	key4 := []byte("4")

	db.Put(key1, []byte("1"))
	db.Put(key2, []byte("2"))

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Put(key1, []byte("a")); err != nil {
		t.Fatal(err)
	}

	if err := tx.Put(key2, []byte("b")); err != nil {
		t.Fatal(err)
	}

	if err := tx.Put(key3, []byte("c")); err != nil {
		t.Fatal(err)
	}

	if err := tx.Put(key4, []byte("d")); err != nil {
		t.Fatal(err)
	}

	it := tx.NewIterator()

	it.Seek(key1)

	if !it.Valid() {
		t.Fatal("must valid")
	} else if string(it.Value()) != "a" {
		t.Fatal(string(it.Value()))
	}

	it.SeekToFirst()

	if !it.Valid() {
		t.Fatal("must valid")
	} else if string(it.Value()) != "a" {
		t.Fatal(string(it.Value()))
	}

	it.Seek(key2)

	if !it.Valid() {
		t.Fatal("must valid")
	} else if string(it.Value()) != "b" {
		t.Fatal(string(it.Value()))
	}

	it.Next()

	if !it.Valid() {
		t.Fatal("must valid")
	} else if string(it.Value()) != "c" {
		t.Fatal(string(it.Value()))
	}

	it.SeekToLast()

	if !it.Valid() {
		t.Fatal("must valid")
	} else if string(it.Value()) != "d" {
		t.Fatal(string(it.Value()))
	}

	it.Close()

	tx.Rollback()

	if v, err := db.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "1" {
		t.Fatal(string(v))
	}

	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Put(key1, []byte("a")); err != nil {
		t.Fatal(err)
	}

	it = tx.NewIterator()

	it.Seek(key2)

	if !it.Valid() {
		t.Fatal("must valid")
	} else if string(it.Value()) != "2" {
		t.Fatal(string(it.Value()))
	}

	it.Close()

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "a" {
		t.Fatal(string(v))
	}
}
