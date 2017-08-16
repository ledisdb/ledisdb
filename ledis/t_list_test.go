package ledis

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
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

func TestListTrim(t *testing.T) {
	db := getTestDB()

	key := []byte("test_list_trim")

	init := func() {
		db.LClear(key)
		for i := 0; i < 100; i++ {
			n, err := db.RPush(key, []byte(strconv.Itoa(i)))
			if err != nil {
				t.Fatal(err)
			}
			if n != int64(i+1) {
				t.Fatal("length wrong")
			}
		}
	}

	init()

	err := db.LTrim(key, 0, 99)
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := db.LLen(key); l != int64(100) {
		t.Fatal("wrong len:", l)
	}

	err = db.LTrim(key, 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := db.LLen(key); l != int64(51) {
		t.Fatal("wrong len:", l)
	}
	for i := int32(0); i < 51; i++ {
		v, err := db.LIndex(key, i)
		if err != nil {
			t.Fatal(err)
		}
		if string(v) != strconv.Itoa(int(i)) {
			t.Fatal("wrong value")
		}
	}

	err = db.LTrim(key, 11, 30)
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := db.LLen(key); l != int64(30-11+1) {
		t.Fatal("wrong len:", l)
	}
	for i := int32(11); i < 31; i++ {
		v, err := db.LIndex(key, i-11)
		if err != nil {
			t.Fatal(err)
		}
		if string(v) != strconv.Itoa(int(i)) {
			t.Fatal("wrong value")
		}
	}

	err = db.LTrim(key, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := db.LLen(key); l != int64(30-11+1) {
		t.Fatal("wrong len:", l)
	}

	init()
	err = db.LTrim(key, -3, -3)
	if err != nil {
		t.Fatal(err)
	}
	if l, _ := db.LLen(key); l != int64(1) {
		t.Fatal("wrong len:", l)
	}
	v, err := db.LIndex(key, 0)
	if err != nil {
		t.Fatal(err)
	}
	if string(v) != "97" {
		t.Fatal("wrong value", string(v))
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

	if ay, err := db.LRange(key, 0, -1); err != nil {
		t.Fatal(err)
	} else if len(ay) != 3 {
		t.Fatal(len(ay))
	} else {
		for i := range ay {
			if ay[i][0] != '1'+byte(i) {
				t.Fatal(string(ay[i]))
			}
		}
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

func TestListPersist(t *testing.T) {
	db := getTestDB()

	key := []byte("persist")
	db.LPush(key, []byte("a"))

	if n, err := db.LPersist(key); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if _, err := db.LExpire(key, 10); err != nil {
		t.Fatal(err)
	}

	if n, err := db.LPersist(key); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}
}

func TestLBlock(t *testing.T) {
	db := getTestDB()

	key1 := []byte("test_lblock_key1")
	key2 := []byte("test_lblock_key2")

	var wg sync.WaitGroup
	wg.Add(2)

	f := func(i int) {
		defer wg.Done()

		ay, err := db.BLPop([][]byte{key1, key2}, 0)
		if err != nil {
			t.Fatal(err)
		} else if len(ay) != 2 {
			t.Fatal(len(ay))
		}
	}

	go f(1)
	go f(2)

	time.Sleep(1 * time.Millisecond)

	db.LPush(key1, []byte("value"))
	db.LPush(key2, []byte("value"))
	wg.Wait()
}

func TestLBlockTimeout(t *testing.T) {
	db := getTestDB()

	key1 := []byte("test_lblock_key1")
	key2 := []byte("test_lblock_key2")

	ay, err := db.BLPop([][]byte{key1, key2}, 10*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	} else if len(ay) != 0 {
		t.Fatal(len(ay))
	}
}

func TestLFlush(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("%d", i)
		if _, err := db.LPush([]byte(key), []byte("v")); err != nil {
			t.Fatal(err.Error())
		}
	}

	if v, err := db.Scan(LIST, nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 2000 {
		t.Fatal("invalid value ", len(v))
	}

	if n, err := db.lFlush(); err != nil {
		t.Fatal(err.Error())
	} else if n != 2000 {
		t.Fatal("invalid value ", n)
	}

	if v, err := db.Scan(LIST, nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 0 {
		t.Fatal("invalid value length ", len(v))
	}
}

func TestLKeyExists(t *testing.T) {
	db := getTestDB()
	key := []byte("lkeyexists_test")
	if n, err := db.LKeyExists(key); err != nil {
		t.Fatal(err.Error())
	} else if n != 0 {
		t.Fatal("invalid value ", n)
	}
	db.LPush(key, []byte("hello"), []byte("world"))
	if n, err := db.LKeyExists(key); err != nil {
		t.Fatal(err.Error())
	} else if n != 1 {
		t.Fatal("invalid value ", n)
	}
}

func TestListPop(t *testing.T) {
	db := getTestDB()

	key := []byte("lpop_test")

	if v, err := db.LPop(key); err != nil {
		t.Fatal(err)
	} else if v != nil {
		t.Fatal(v)
	}

	if s, err := db.LLen(key); err != nil {
		t.Fatal(err)
	} else if s != 0 {
		t.Fatal(s)
	}

	for i := 0; i < 10; i++ {
		if n, err := db.LPush(key, []byte("a")); err != nil {
			t.Fatal(err)
		} else if n != 1+int64(i) {
			t.Fatal(n)
		}
	}

	if s, err := db.LLen(key); err != nil {
		t.Fatal(err)
	} else if s != 10 {
		t.Fatal(s)
	}

	for i := 0; i < 10; i++ {
		if _, err := db.LPop(key); err != nil {
			t.Fatal(err)
		}
	}

	if s, err := db.LLen(key); err != nil {
		t.Fatal(err)
	} else if s != 0 {
		t.Fatal(s)
	}

	if v, err := db.LPop(key); err != nil {
		t.Fatal(err)
	} else if v != nil {
		t.Fatal(v)
	}

}
