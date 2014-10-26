package ledis

import (
	"fmt"
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
func TestKVFlush(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("%d", i)
		if err := db.Set([]byte(key), []byte("v")); err != nil {
			t.Fatal(err.Error())
		}
	}

	if v, err := db.Scan(nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 2000 {
		t.Fatal("invalid value ", len(v))
	}

	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("%d", i)
		if v, err := db.Get([]byte(key)); err != nil {
			t.Fatal(err.Error())
		} else if string(v) != "v" {
			t.Fatal("invalid value ", v)
		}
	}

	if n, err := db.flush(); err != nil {
		t.Fatal(err.Error())
	} else if n != 2000 {
		t.Fatal("invalid value ", n)
	}

	if v, err := db.Scan(nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 0 {
		t.Fatal("invalid value length ", len(v))
	}

	for i := 0; i < 2000; i++ {

		key := []byte(fmt.Sprintf("%d", i))

		if v, err := db.Get(key); err != nil {
			t.Fatal(err.Error())
		} else if v != nil {

			t.Fatal("invalid value ", v)
		}
	}
}

func TestKVSetEX(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	key := []byte("testdb_kv_c")

	if err := db.SetEX(key, 10, []byte("hello world")); err != nil {
		t.Fatal(err)
	}

	v, err := db.Get(key)
	if err != nil {
		t.Fatal(err)
	} else if string(v) == "" {
		t.Fatal("v is nil")
	}

	if n, err := db.TTL(key); err != nil {
		t.Fatal(err)
	} else if n != 10 {
		t.Fatal(n)
	}

	if v, _ := db.Get(key); string(v) != "hello world" {
		t.Fatal(string(v))
	}

}
