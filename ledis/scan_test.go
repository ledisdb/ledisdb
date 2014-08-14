package ledis

import (
	"testing"
)

func TestDBScan(t *testing.T) {
	db := getTestDB()

	db.FlushAll()

	if v, err := db.Scan(nil, 10, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(len(v))
	}

	db.Set([]byte("a"), []byte{})
	db.Set([]byte("b"), []byte{})
	db.Set([]byte("c"), []byte{})

	if v, err := db.Scan(nil, 1, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal(len(v))
	}

	if v, err := db.Scan([]byte("a"), 2, false); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(len(v))
	}

	if v, err := db.Scan(nil, 3, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 3 {
		t.Fatal(len(v))
	}
}

func TestDBHScan(t *testing.T) {
	db := getTestDB()

	db.hFlush()

	k1 := []byte("k1")
	db.HSet(k1, []byte("1"), []byte{})

	k2 := []byte("k2")
	db.HSet(k2, []byte("2"), []byte{})

	k3 := []byte("k3")
	db.HSet(k3, []byte("3"), []byte{})

	if v, err := db.HScan(nil, 1, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.HScan(k1, 2, true); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.HScan(k1, 2, false); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}
