package ledis

import (
	"testing"
)

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

	key := []byte("testdb_zset_a")

	if n, err := db.ZAdd(key, ScorePair{1, []byte("a")}, ScorePair{1, []byte("a")}); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}
}
