package ledis

import (
	"testing"
)

func TestDBZSet(t *testing.T) {
	db := getTestDB()

	key := []byte("testdb_zset_a")

	if n, err := db.ZAdd(key, ScorePair{1, []byte("a")}, ScorePair{1, []byte("a")}); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}
}
