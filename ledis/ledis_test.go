package ledis

import (
	"os"
	"sync"
	"testing"

	"github.com/siddontang/ledisdb/config"
)

var testLedis *Ledis
var testLedisOnce sync.Once

func getTestDB() *DB {
	f := func() {
		cfg := config.NewConfigDefault()
		cfg.DataDir = "/tmp/test_ledis"
		cfg.Databases = 10240

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
	db1024, _ := testLedis.Select(1024)

	testSelect(t, db0)
	testSelect(t, db1)
	testSelect(t, db1024)
}

func testSelect(t *testing.T, db *DB) {
	key := []byte("test_select_key")
	value := []byte("value")
	if err := db.Set(key, value); err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get(key); err != nil {
		t.Fatal(err)
	} else if string(v) != string(value) {
		t.Fatal(string(v))
	}

	if _, err := db.Expire(key, 100); err != nil {
		t.Fatal(err)
	}

	if _, err := db.TTL(key); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Persist(key); err != nil {
		t.Fatal(err)
	}

	key = []byte("test_select_list_key")
	if _, err := db.LPush(key, value); err != nil {
		t.Fatal(err)
	}

	if _, err := db.LRange(key, 0, 100); err != nil {
		t.Fatal(err)
	}

	if v, err := db.LPop(key); err != nil {
		t.Fatal(err)
	} else if string(v) != string(value) {
		t.Fatal(string(v))
	}

	key = []byte("test_select_hash_key")
	if _, err := db.HSet(key, []byte("a"), value); err != nil {
		t.Fatal(err)
	}

	if v, err := db.HGet(key, []byte("a")); err != nil {
		t.Fatal(err)
	} else if string(v) != string(value) {
		t.Fatal(string(v))
	}

	key = []byte("test_select_set_key")
	if _, err := db.SAdd(key, []byte("a"), []byte("b")); err != nil {
		t.Fatal(err)
	}

	if n, err := db.SIsMember(key, []byte("a")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	key = []byte("test_select_zset_key")
	if _, err := db.ZAdd(key, ScorePair{1, []byte("a")}, ScorePair{2, []byte("b")}); err != nil {
		t.Fatal(err)
	}

	if v, err := db.ZRangeByScore(key, 0, 100, 0, -1); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(len(v))
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
