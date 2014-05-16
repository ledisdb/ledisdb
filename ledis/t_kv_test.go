package ledis

import (
	"testing"
)

func TestDBKV(t *testing.T) {
	db := getTestDB()

	key := []byte("testdb_kv_a")

	if err := db.Set(key, []byte("hello world")); err != nil {
		t.Fatal(err)
	}
}
