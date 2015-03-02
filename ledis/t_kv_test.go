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

	key3 := []byte("testdb_kv_range")

	if n, err := db.Append(key3, []byte("Hello")); err != nil {
		t.Fatal(err)
	} else if n != 5 {
		t.Fatal(n)
	}

	if n, err := db.Append(key3, []byte(" World")); err != nil {
		t.Fatal(err)
	} else if n != 11 {
		t.Fatal(n)
	}

	if n, err := db.StrLen(key3); err != nil {
		t.Fatal(err)
	} else if n != 11 {
		t.Fatal(n)
	}

	if v, err := db.GetRange(key3, 0, 4); err != nil {
		t.Fatal(err)
	} else if string(v) != "Hello" {
		t.Fatal(string(v))
	}

	if v, err := db.GetRange(key3, 0, -1); err != nil {
		t.Fatal(err)
	} else if string(v) != "Hello World" {
		t.Fatal(string(v))
	}

	if v, err := db.GetRange(key3, -5, -1); err != nil {
		t.Fatal(err)
	} else if string(v) != "World" {
		t.Fatal(string(v))
	}

	if n, err := db.SetRange(key3, 6, []byte("Redis")); err != nil {
		t.Fatal(err)
	} else if n != 11 {
		t.Fatal(n)
	}

	if v, err := db.Get(key3); err != nil {
		t.Fatal(err)
	} else if string(v) != "Hello Redis" {
		t.Fatal(string(v))
	}

	key4 := []byte("testdb_kv_range_none")
	if n, err := db.SetRange(key4, 6, []byte("Redis")); err != nil {
		t.Fatal(err)
	} else if n != 11 {
		t.Fatal(n)
	}

	key5 := []byte("testdb_kv_bit")
	if n, err := db.SetBit(key5, 7, 1); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := db.GetBit(key5, 0); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}
	if n, err := db.GetBit(key5, 7); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}
	if n, err := db.GetBit(key5, 100); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := db.BitCount(key5, 0, -1); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := db.BitPos(key5, 1, 0, -1); err != nil {
		t.Fatal(err)
	} else if n != 7 {
		t.Fatal(n)
	}

	if n, err := db.SetBit(key5, 7, 0); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if err := db.Set(key5, []byte{0xff, 0xf0, 0x00}); err != nil {
		t.Fatal(err)
	}

	if n, err := db.BitPos(key5, 0, 0, -1); err != nil {
		t.Fatal(err)
	} else if n != 12 {
		t.Fatal(n)
	}

	if err := db.Set(key5, []byte{0x00, 0xff, 0xf0}); err != nil {
		t.Fatal(err)
	}

	if n, err := db.BitPos(key5, 1, 0, -1); err != nil {
		t.Fatal(err)
	} else if n != 8 {
		t.Fatal(n)
	}

	if n, err := db.BitPos(key5, 1, 2, -1); err != nil {
		t.Fatal(err)
	} else if n != 16 {
		t.Fatal(n)
	}

	if err := db.Set(key5, []byte{0x00, 0x00, 0x00}); err != nil {
		t.Fatal(err)
	}

	if n, err := db.BitPos(key5, 1, 0, -1); err != nil {
		t.Fatal(err)
	} else if n != -1 {
		t.Fatal(n)
	}

	if err := db.Set(key5, []byte("foobar")); err != nil {
		t.Fatal(err)
	}

	if n, err := db.BitCount(key5, 0, -1); err != nil {
		t.Fatal(err)
	} else if n != 26 {
		t.Fatal(n)
	}

	if n, err := db.BitCount(key5, 0, 0); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if n, err := db.BitCount(key5, 1, 1); err != nil {
		t.Fatal(err)
	} else if n != 6 {
		t.Fatal(n)
	}

	key6 := []byte("testdb_kv_bitop_desc")

	if err := db.Set(key4, []byte("foobar")); err != nil {
		t.Fatal(err)
	}

	if err := db.Set(key5, []byte("abcdef")); err != nil {
		t.Fatal(err)
	}

	if n, err := db.BitOP(BitAND, key6, key4, key5); err != nil {
		t.Fatal(err)
	} else if n != 6 {
		t.Fatal(err)
	}

	if v, err := db.Get(key6); err != nil {
		t.Fatal(err)
	} else if string(v) != "`bc`ab" {
		t.Fatal(string(v))
	}

	if err := db.Set(key4, []byte("fooba")); err != nil {
		t.Fatal(err)
	}

	if n, err := db.BitOP(BitOR, key6, key4, key5); err != nil {
		t.Fatal(err)
	} else if n != 6 {
		t.Fatal(err)
	}

	if v, err := db.Get(key6); err != nil {
		t.Fatal(err)
	} else if string(v) != "goofef" {
		t.Fatal(string(v))
	}

	if n, err := db.BitOP(BitAND, key6, key4, key5); err != nil {
		t.Fatal(err)
	} else if n != 6 {
		t.Fatal(err)
	}

	if v, err := db.Get(key6); err != nil {
		t.Fatal(err)
	} else if string(v) != "`bc`a\x00" {
		t.Fatal(string(v))
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

	if v, err := db.Scan(KV, nil, 3000, true, ""); err != nil {
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

	if v, err := db.Scan(KV, nil, 3000, true, ""); err != nil {
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
