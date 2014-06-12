package ledis

import (
	"testing"
)

func TestKVCodec(t *testing.T) {
	db := getTestDB()

	ek := db.encodeKVKey([]byte("key"))

	if k, err := db.decodeKVKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	}
}

func TestDBKV(t *testing.T) {
	db := getTestDB()

	key := []byte("testdb_kv_a")

	if err := db.Set(key, []byte("hello world")); err != nil {
		t.Fatal(err)
	}
}

func TestDBScan(t *testing.T) {
	db := getTestDB()

	db.FlushAll()

	if v, err := db.Scan(nil, 10, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(len(v))
	}

	db.Set([]byte("a"), []byte{})
	db.Set([]byte("b"), []byte{})
	db.Set([]byte("c"), []byte{})

	if v, err := db.Scan(nil, 1, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal(len(v))
	}

	if v, err := db.Scan([]byte("a"), 2, false); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(len(v))
	}

	if v, err := db.Scan(nil, 3, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 3 {
		t.Fatal(len(v))
	}
}
