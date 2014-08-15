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

	key1 := []byte("testdb_kv_a")

	if err := db.Set(key1, []byte("hello world 1")); err != nil {
		t.Fatal(err)
	}

	key2 := []byte("testdb_kv_b")

	if err := db.Set(key2, []byte("hello world 2")); err != nil {
		t.Fatal(err)
	}

	ay, _ := db.MGet(key1, key2)

	if v1 := ay[0]; string(v1) != "hello world 1" {
		t.Fatal(string(v1))
	}

	if v2 := ay[1]; string(v2) != "hello world 2" {
		t.Fatal(string(v2))
	}

}

func TestKVPersist(t *testing.T) {
	db := getTestDB()

	key := []byte("persist")
	db.Set(key, []byte{})

	if n, err := db.Persist(key); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if _, err := db.Expire(key, 10); err != nil {
		t.Fatal(err)
	}

	if n, err := db.Persist(key); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}
}
