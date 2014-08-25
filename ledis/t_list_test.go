package ledis

import (
	"fmt"
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

func TestLFlush(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("%d", i)
		if _, err := db.LPush([]byte(key), []byte("v")); err != nil {
			t.Fatal(err.Error())
		}
	}

	if v, err := db.LScan(nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 2000 {
		t.Fatal("invalid value ", len(v))
	}

	if n, err := db.lFlush(); err != nil {
		t.Fatal(err.Error())
	} else if n != 2000 {
		t.Fatal("invalid value ", n)
	}

	if v, err := db.LScan(nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 0 {
		t.Fatal("invalid value length ", len(v))
	}
}
