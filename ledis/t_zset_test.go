package ledis

import (
	"fmt"
	"testing"
)

const (
	endPos int = -1
)

func bin(sz string) []byte {
	return []byte(sz)
}

func pair(memb string, score int) ScorePair {
	return ScorePair{int64(score), bin(memb)}
}

func TestZSetCodec(t *testing.T) {
	db := getTestDB()

	key := []byte("key")
	member := []byte("member")

	ek := db.zEncodeSizeKey(key)
	if k, err := db.zDecodeSizeKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	}

	ek = db.zEncodeSetKey(key, member)
	if k, m, err := db.zDecodeSetKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	} else if string(m) != "member" {
		t.Fatal(string(m))
	}

	ek = db.zEncodeScoreKey(key, member, 100)
	if k, m, s, err := db.zDecodeScoreKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	} else if string(m) != "member" {
		t.Fatal(string(m))
	} else if s != 100 {
		t.Fatal(s)
	}

}

func TestDBZSet(t *testing.T) {
	db := getTestDB()

	key := bin("testdb_zset_a")

	// {'a':0, 'b':1, 'c':2, 'd':3}
	if n, err := db.ZAdd(key, pair("a", 0), pair("b", 1),
		pair("c", 2), pair("d", 3)); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if n, err := db.ZCount(key, 0, 0XFF); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if s, err := db.ZScore(key, bin("d")); err != nil {
		t.Fatal(err)
	} else if s != 3 {
		t.Fatal(s)
	}

	if s, err := db.ZScore(key, bin("zzz")); err != ErrScoreMiss || s != InvalidScore {
		t.Fatal(fmt.Sprintf("s=[%d] err=[%s]", s, err))
	}

	// {c':2, 'd':3}
	if n, err := db.ZRem(key, bin("a"), bin("b")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := db.ZRem(key, bin("a"), bin("b")); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := db.ZCard(key); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	// {}
	if n, err := db.ZClear(key); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := db.ZCount(key, 0, 0XFF); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}
}

func TestZSetOrder(t *testing.T) {
	db := getTestDB()

	key := bin("testdb_zset_order")

	// {'a':0, 'b':1, 'c':2, 'd':3, 'e':4, 'f':5}
	membs := [...]string{"a", "b", "c", "d", "e", "f"}
	membCnt := len(membs)

	for i := 0; i < membCnt; i++ {
		db.ZAdd(key, pair(membs[i], i))
	}

	if n, _ := db.ZCount(key, 0, 0XFFFF); int(n) != membCnt {
		t.Fatal(n)
	}

	for i := 0; i < membCnt; i++ {
		if pos, err := db.ZRank(key, bin(membs[i])); err != nil {
			t.Fatal(err)
		} else if int(pos) != i {
			t.Fatal(pos)
		}

		if pos, err := db.ZRevRank(key, bin(membs[i])); err != nil {
			t.Fatal(err)
		} else if int(pos) != membCnt-i-1 {
			t.Fatal(pos)
		}
	}

	if qMembs, err := db.ZRange(key, 0, endPos); err != nil {
		t.Fatal(err)
	} else if len(qMembs) != membCnt {
		t.Fatal(fmt.Sprintf("%d vs %d", len(qMembs), membCnt))
	} else {
		for i := 0; i < membCnt; i++ {
			if string(qMembs[i].Member) != membs[i] {
				t.Fatal(fmt.Sprintf("[%s] vs [%s]", qMembs[i], membs[i]))
			}
		}
	}

	// {'a':0, 'b':1, 'c':2, 'd':999, 'e':4, 'f':5}
	if n, err := db.ZAdd(key, pair("d", 999)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if pos, _ := db.ZRank(key, bin("d")); int(pos) != membCnt-1 {
		t.Fatal(pos)
	}

	if pos, _ := db.ZRevRank(key, bin("d")); int(pos) != 0 {
		t.Fatal(pos)
	}

	if pos, _ := db.ZRank(key, bin("e")); int(pos) != 3 {
		t.Fatal(pos)
	}

	if pos, _ := db.ZRank(key, bin("f")); int(pos) != 4 {
		t.Fatal(pos)
	}

	if qMembs, err := db.ZRangeByScore(key, 999, 0XFFFF, 0, membCnt); err != nil {
		t.Fatal(err)
	} else if len(qMembs) != 1 {
		t.Fatal(len(qMembs))
	}

	// {'a':0, 'b':1, 'c':2, 'd':999, 'e':6, 'f':5}
	if s, err := db.ZIncrBy(key, 2, bin("e")); err != nil {
		t.Fatal(err)
	} else if s != 6 {
		t.Fatal(s)
	}

	if pos, _ := db.ZRank(key, bin("e")); int(pos) != 4 {
		t.Fatal(pos)
	}

	if pos, _ := db.ZRevRank(key, bin("e")); int(pos) != 1 {
		t.Fatal(pos)
	}

	if datas, _ := db.ZRange(key, 0, endPos); len(datas) != 6 {
		t.Fatal(len(datas))
	} else {
		scores := []int64{0, 1, 2, 5, 6, 999}
		for i := 0; i < len(datas); i++ {
			if datas[i].Score != scores[i] {
				t.Fatal(fmt.Sprintf("[%d]=%d", i, datas[i]))
			}
		}
	}

	return
}

func TestDBZScan(t *testing.T) {
	db := getTestDB()

	db.zFlush()

	key := []byte("key")
	db.ZAdd(key, pair("a", 0), pair("b", 1), pair("c", 2))

	if v, err := db.ZScan(key, nil, 1, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal(len(v))
	}

	if v, err := db.ZScan(key, []byte("a"), 2, false); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(len(v))
	}

	if v, err := db.ZScan(key, nil, 10, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 3 {
		t.Fatal(len(v))
	}
}

func TestZSetPersist(t *testing.T) {
	db := getTestDB()

	key := []byte("persist")
	db.ZAdd(key, ScorePair{1, []byte("a")})

	if n, err := db.ZPersist(key); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if _, err := db.ZExpire(key, 10); err != nil {
		t.Fatal(err)
	}

	if n, err := db.ZPersist(key); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}
}

func TestZUnionStore(t *testing.T) {
	db := getTestDB()
	key1 := []byte("key1")
	key2 := []byte("key2")

	db.ZAdd(key1, ScorePair{1, []byte("one")})
	db.ZAdd(key1, ScorePair{1, []byte("two")})

	db.ZAdd(key2, ScorePair{2, []byte("two")})
	db.ZAdd(key2, ScorePair{2, []byte("three")})

	keys := [][]byte{key1, key2}
	weights := []int64{1, 2}

	out := []byte("out")
	n, err := db.ZUnionStore(out, keys, weights, AggregateSum)
	if err != nil {
		t.Fatal(err.Error())
	}
	if n != 3 {
		t.Fatal("invalid value ", n)
	}

	v, err := db.ZScore(out, []byte("two"))

	if err != nil {
		t.Fatal(err.Error())
	}
	if v != 5 {
		t.Fatal("invalid value ", v)
	}
}
