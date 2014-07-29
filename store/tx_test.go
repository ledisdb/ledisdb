package store

import (
	"testing"
)

func TestTx(t *testing.T) {

}

func testTx(db *DB, t *testing.T) {
	key1 := []byte("1")
	key2 := []byte("2")

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

	it := tx.NewIterator()

	it.Seek(key1)

	if !it.Valid() {
		t.Fatal("must valid")
	} else if string(it.Value()) != "a" {
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
