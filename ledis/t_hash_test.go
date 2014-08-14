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

	if v1 := ay[0]; string(v1) != "hello world 1" {
		t.Fatal(string(v1))
	}

	if v2 := ay[1]; string(v2) != "hello world 2" {
		t.Fatal(string(v2))
	}

}

func TestHashPersist(t *testing.T) {
	db := getTestDB()

	key := []byte("persist")
	db.HSet(key, []byte("field"), []byte{})

	if n, err := db.HPersist(key); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if _, err := db.HExpire(key, 10); err != nil {
		t.Fatal(err)
	}

	if n, err := db.HPersist(key); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}
}
