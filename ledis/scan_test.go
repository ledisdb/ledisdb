package ledis

import (
	"testing"
)

func checkTestScan(t *testing.T, v [][]byte, args ...string) {
	if len(v) != len(args) {
		t.Fatal(len(v), len(args))
	}

	for i := range v {
		if string(v[i]) != args[i] {
			t.Fatalf("%q %q", v, args)
		}
	}
}

func TestDBScan(t *testing.T) {
	db := getTestDB()

	db.FlushAll()

	if v, err := db.Scan(nil, 10, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(len(v))
	}

	if v, err := db.RevScan(nil, 10, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(len(v))
	}

	db.Set([]byte("a"), []byte{})
	db.Set([]byte("b"), []byte{})
	db.Set([]byte("c"), []byte{})

	if v, err := db.Scan(nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "a")
	}

	if v, err := db.Scan([]byte("a"), 2, false, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "b", "c")
	}

	if v, err := db.Scan(nil, 3, true, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "a", "b", "c")
	}

	if v, err := db.Scan(nil, 3, true, "b"); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "b")
	}

	if v, err := db.Scan(nil, 3, true, "."); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "a", "b", "c")
	}

	if v, err := db.Scan(nil, 3, true, "a+"); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "a")
	}

	if v, err := db.RevScan(nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "c")
	}

	if v, err := db.RevScan([]byte("c"), 2, false, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "b", "a")
	}

	if v, err := db.RevScan(nil, 3, true, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "c", "b", "a")
	}

	if v, err := db.RevScan(nil, 3, true, "b"); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "b")
	}

	if v, err := db.RevScan(nil, 3, true, "."); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "c", "b", "a")
	}

	if v, err := db.RevScan(nil, 3, true, "c+"); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "c")
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

	if v, err := db.HScan(nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.HScan(k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.HScan(k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}

func TestDBZScan(t *testing.T) {
	db := getTestDB()

	db.zFlush()

	k1 := []byte("k1")
	db.ZAdd(k1, ScorePair{1, []byte("m")})

	k2 := []byte("k2")
	db.ZAdd(k2, ScorePair{2, []byte("m")})

	k3 := []byte("k3")
	db.ZAdd(k3, ScorePair{3, []byte("m")})

	if v, err := db.ZScan(nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.ZScan(k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.ZScan(k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}

func TestDBLScan(t *testing.T) {
	db := getTestDB()

	db.lFlush()

	k1 := []byte("k1")
	if _, err := db.LPush(k1, []byte("elem")); err != nil {
		t.Fatal(err.Error())
	}

	k2 := []byte("k2")
	if _, err := db.LPush(k2, []byte("elem")); err != nil {
		t.Fatal(err.Error())
	}

	k3 := []byte("k3")
	if _, err := db.LPush(k3, []byte("elem")); err != nil {
		t.Fatal(err.Error())
	}

	if v, err := db.LScan(nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.LScan(k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.LScan(k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}

func TestDBBScan(t *testing.T) {
	db := getTestDB()

	db.bFlush()

	k1 := []byte("k1")
	if _, err := db.BSetBit(k1, 1, 1); err != nil {
		t.Fatal(err.Error())
	}

	k2 := []byte("k2")
	if _, err := db.BSetBit(k2, 1, 1); err != nil {
		t.Fatal(err.Error())
	}
	k3 := []byte("k3")

	if _, err := db.BSetBit(k3, 1, 0); err != nil {
		t.Fatal(err.Error())
	}

	if v, err := db.BScan(nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.BScan(k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.BScan(k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}

func TestDBSScan(t *testing.T) {
	db := getTestDB()

	db.bFlush()

	k1 := []byte("k1")
	if _, err := db.SAdd(k1, []byte("1")); err != nil {
		t.Fatal(err.Error())
	}

	k2 := []byte("k2")
	if _, err := db.SAdd(k2, []byte("1")); err != nil {
		t.Fatal(err.Error())
	}
	k3 := []byte("k3")

	if _, err := db.SAdd(k3, []byte("1")); err != nil {
		t.Fatal(err.Error())
	}

	if v, err := db.SScan(nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.SScan(k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.SScan(k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}
