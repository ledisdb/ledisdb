package ledis

import (
	"github.com/siddontang/ledisdb/config"
	"os"
	"testing"
)

func testTxRollback(t *testing.T, db *DB) {
	var err error
	key1 := []byte("tx_key1")
	key2 := []byte("tx_key2")
	field2 := []byte("tx_field2")

	err = db.Set(key1, []byte("value"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.HSet(key2, field2, []byte("value"))
	if err != nil {
		t.Fatal(err)
	}

	var tx *Tx
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	defer tx.Rollback()

	err = tx.Set(key1, []byte("1"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.HSet(key2, field2, []byte("2"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.HSet([]byte("no_key"), field2, []byte("2"))

	if err != nil {
		t.Fatal(err)
	}

	if v, err := tx.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "1" {
		t.Fatal(string(v))
	}

	if v, err := tx.HGet(key2, field2); err != nil {
		t.Fatal(err)
	} else if string(v) != "2" {
		t.Fatal(string(v))
	}

	err = tx.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "value" {
		t.Fatal(string(v))
	}

	if v, err := db.HGet(key2, field2); err != nil {
		t.Fatal(err)
	} else if string(v) != "value" {
		t.Fatal(string(v))
	}
}

func testTxCommit(t *testing.T, db *DB) {
	var err error
	key1 := []byte("tx_key1")
	key2 := []byte("tx_key2")
	field2 := []byte("tx_field2")

	err = db.Set(key1, []byte("value"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.HSet(key2, field2, []byte("value"))
	if err != nil {
		t.Fatal(err)
	}

	var tx *Tx
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	defer tx.Rollback()

	err = tx.Set(key1, []byte("1"))

	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.HSet(key2, field2, []byte("2"))

	if err != nil {
		t.Fatal(err)
	}

	if v, err := tx.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "1" {
		t.Fatal(string(v))
	}

	if v, err := tx.HGet(key2, field2); err != nil {
		t.Fatal(err)
	} else if string(v) != "2" {
		t.Fatal(string(v))
	}

	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "1" {
		t.Fatal(string(v))
	}

	if v, err := db.HGet(key2, field2); err != nil {
		t.Fatal(err)
	} else if string(v) != "2" {
		t.Fatal(string(v))
	}
}

func testTx(t *testing.T, name string) {
	cfg := new(config.Config)
	cfg.DataDir = "/tmp/ledis_test_tx"

	cfg.DBName = name
	cfg.LMDB.MapSize = 10 * 1024 * 1024

	os.RemoveAll(cfg.DataDir)

	l, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}

	defer l.Close()

	db, _ := l.Select(0)

	testTxRollback(t, db)
	testTxCommit(t, db)
}

//only lmdb, boltdb support Transaction
func TestTx(t *testing.T) {
	testTx(t, "lmdb")
	testTx(t, "boltdb")
}
