package ledis

import (
	"testing"
)

func TestDBList(t *testing.T) {
	db := getTestDB()

	key := []byte("testdb_list_a")

	if n, err := db.RPush(key, []byte("1"), []byte("2")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}
}
