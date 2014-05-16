package ledis

import (
	"testing"
)

func TestDBHash(t *testing.T) {
	db := getTestDB()

	key := []byte("testdb_hash_a")

	if n, err := db.HSet(key, []byte("a"), []byte("hello world")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}
}
