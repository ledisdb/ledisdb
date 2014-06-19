package ledis

import (
	"testing"
)

func TestHashCodec(t *testing.T) {
	db := getTestDB()

	key := []byte("key")
	field := []byte("field")

	ek := db.hEncodeSizeKey(key)
	if k, err := db.hDecodeSizeKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	}

	ek = db.hEncodeHashKey(key, field)
	if k, f, err := db.hDecodeHashKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	} else if string(f) != "field" {
		t.Fatal(string(f))
	}
}

func TestDBHash(t *testing.T) {
	db := getTestDB()

	key := []byte("testdb_hash_a")

	if n, err := db.HSet(key, []byte("a"), []byte("hello world 1")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := db.HSet(key, []byte("b"), []byte("hello world 2")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	ay, _ := db.HMget(key, []byte("a"), []byte("b"))

	if v1, _ := ay[0].([]byte); string(v1) != "hello world 1" {
		t.Fatal(string(v1))
	}

	if v2, _ := ay[1].([]byte); string(v2) != "hello world 2" {
		t.Fatal(string(v2))
	}

}

func TestDBHScan(t *testing.T) {
	db := getTestDB()

	db.hFlush()

	key := []byte("a")
	db.HSet(key, []byte("1"), []byte{})
	db.HSet(key, []byte("2"), []byte{})
	db.HSet(key, []byte("3"), []byte{})

	if v, err := db.HScan(key, nil, 1, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal(len(v))
	}

	if v, err := db.HScan(key, []byte("1"), 2, false); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(len(v))
	}

	if v, err := db.HScan(key, nil, 10, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 3 {
		t.Fatal(len(v))
	}

}
