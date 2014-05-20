package ledis

import (
	"testing"
)

func TestListCodec(t *testing.T) {
	db := getTestDB()

	key := []byte("key")

	ek := db.lEncodeMetaKey(key)
	if k, err := db.lDecodeMetaKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	}

	ek = db.lEncodeListKey(key, 1024)
	if k, seq, err := db.lDecodeListKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	} else if seq != 1024 {
		t.Fatal(seq)
	}
}

func TestDBList(t *testing.T) {
	db := getTestDB()

	key := []byte("testdb_list_a")

	if n, err := db.RPush(key, []byte("1"), []byte("2")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}
}
