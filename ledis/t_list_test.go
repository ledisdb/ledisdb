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

	if n, err := db.RPush(key, []byte("1"), []byte("2"), []byte("3")); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if k, err := db.RPop(key); err != nil {
		t.Fatal(err)
	} else if string(k) != "3" {
		t.Fatal(string(k))
	}

	if k, err := db.LPop(key); err != nil {
		t.Fatal(err)
	} else if string(k) != "1" {
		t.Fatal(string(k))
	}

	if llen, err := db.LLen(key); err != nil {
		t.Fatal(err)
	} else if llen != 1 {
		t.Fatal(llen)
	}

	if num, err := db.LClear(key); err != nil {
		t.Fatal(err)
	} else if num != 1 {
		t.Fatal(num)
	}

	if llen, _ := db.LLen(key); llen != 0 {
		t.Fatal(llen)
	}
}
