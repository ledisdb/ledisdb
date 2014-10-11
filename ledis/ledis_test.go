package ledis

import (
	"github.com/siddontang/ledisdb/config"
	"os"
	"sync"
	"testing"
)

var testLedis *Ledis
var testLedisOnce sync.Once

func getTestDB() *DB {
	f := func() {
		cfg := config.NewConfigDefault()
		cfg.DataDir = "/tmp/test_ledis"

		os.RemoveAll(cfg.DataDir)

		var err error
		testLedis, err = Open(cfg)
		if err != nil {
			println(err.Error())
			panic(err)
		}
	}

	testLedisOnce.Do(f)
	db, _ := testLedis.Select(0)
	return db
}

func TestDB(t *testing.T) {
	getTestDB()
}

func TestSelect(t *testing.T) {
	db0, _ := testLedis.Select(0)
	db1, _ := testLedis.Select(1)

	key0 := []byte("db0_test_key")
	key1 := []byte("db1_test_key")

	db0.Set(key0, []byte("0"))
	db1.Set(key1, []byte("1"))

	if v, err := db0.Get(key0); err != nil {
		t.Fatal(err)
	} else if string(v) != "0" {
		t.Fatal(string(v))
	}

	if v, err := db1.Get(key1); err != nil {
		t.Fatal(err)
	} else if string(v) != "1" {
		t.Fatal(string(v))
	}
}

func TestFlush(t *testing.T) {
	db0, _ := testLedis.Select(0)
	db1, _ := testLedis.Select(1)

	db0.Set([]byte("a"), []byte("1"))
	db0.ZAdd([]byte("zset_0"), ScorePair{int64(1), []byte("ma")})
	db0.ZAdd([]byte("zset_0"), ScorePair{int64(2), []byte("mb")})

	db1.Set([]byte("b"), []byte("2"))
	db1.LPush([]byte("lst"), []byte("a1"), []byte("b2"))
	db1.ZAdd([]byte("zset_0"), ScorePair{int64(3), []byte("mc")})

	db1.FlushAll()

	//	0 - existing
	if exists, _ := db0.Exists([]byte("a")); exists <= 0 {
		t.Fatal(false)
	}

	if zcnt, _ := db0.ZCard([]byte("zset_0")); zcnt != 2 {
		t.Fatal(zcnt)
	}

	//	1 - deleted
	if exists, _ := db1.Exists([]byte("b")); exists > 0 {
		t.Fatal(false)
	}

	if llen, _ := db1.LLen([]byte("lst")); llen > 0 {
		t.Fatal(llen)
	}

	if zcnt, _ := db1.ZCard([]byte("zset_1")); zcnt > 0 {
		t.Fatal(zcnt)
	}
}
