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

	if v, err := db.Scan(KV, nil, 10, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(len(v))
	}

	if v, err := db.RevScan(KV, nil, 10, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(len(v))
	}

	db.Set([]byte("a"), []byte{})
	db.Set([]byte("b"), []byte{})
	db.Set([]byte("c"), []byte{})

	if v, err := db.Scan(KV, nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "a")
	}

	if v, err := db.Scan(KV, []byte("a"), 2, false, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "b", "c")
	}

	if v, err := db.Scan(KV, nil, 3, true, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "a", "b", "c")
	}

	if v, err := db.Scan(KV, nil, 3, true, "b"); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "b")
	}

	if v, err := db.Scan(KV, nil, 3, true, "."); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "a", "b", "c")
	}

	if v, err := db.Scan(KV, nil, 3, true, "a+"); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "a")
	}

	if v, err := db.RevScan(KV, nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "c")
	}

	if v, err := db.RevScan(KV, []byte("c"), 2, false, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "b", "a")
	}

	if v, err := db.RevScan(KV, nil, 3, true, ""); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "c", "b", "a")
	}

	if v, err := db.RevScan(KV, nil, 3, true, "b"); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "b")
	}

	if v, err := db.RevScan(KV, nil, 3, true, "."); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "c", "b", "a")
	}

	if v, err := db.RevScan(KV, nil, 3, true, "c+"); err != nil {
		t.Fatal(err)
	} else {
		checkTestScan(t, v, "c")
	}

}

func TestDBHKeyScan(t *testing.T) {
	db := getTestDB()

	db.hFlush()

	k1 := []byte("k1")
	db.HSet(k1, []byte("1"), []byte{})

	k2 := []byte("k2")
	db.HSet(k2, []byte("2"), []byte{})

	k3 := []byte("k3")
	db.HSet(k3, []byte("3"), []byte{})

	if v, err := db.Scan(HASH, nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.Scan(HASH, k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.Scan(HASH, k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}

func TestDBZKeyScan(t *testing.T) {
	db := getTestDB()

	db.zFlush()

	k1 := []byte("k1")
	db.ZAdd(k1, ScorePair{1, []byte("m")})

	k2 := []byte("k2")
	db.ZAdd(k2, ScorePair{2, []byte("m")})

	k3 := []byte("k3")
	db.ZAdd(k3, ScorePair{3, []byte("m")})

	if v, err := db.Scan(ZSET, nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.Scan(ZSET, k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.Scan(ZSET, k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}

func TestDBLKeyScan(t *testing.T) {
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

	if v, err := db.Scan(LIST, nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.Scan(LIST, k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.Scan(LIST, k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}

}

func TestDBSKeyScan(t *testing.T) {
	db := getTestDB()

	db.sFlush()

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

	if v, err := db.Scan(SET, nil, 1, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	}

	if v, err := db.Scan(SET, k1, 2, true, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k1" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k2" {
		t.Fatal("invalid value ", string(v[1]))
	}

	if v, err := db.Scan(SET, k1, 2, false, ""); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal("invalid length ", len(v))
	} else if string(v[0]) != "k2" {
		t.Fatal("invalid value ", string(v[0]))
	} else if string(v[1]) != "k3" {
		t.Fatal("invalid value ", string(v[1]))
	}
}

func TestDBHScan(t *testing.T) {
	db := getTestDB()

	key := []byte("scan_h_key")
	value := []byte("hello world")
	db.HSet(key, []byte("1"), value)
	db.HSet(key, []byte("222"), value)
	db.HSet(key, []byte("19"), value)
	db.HSet(key, []byte("1234"), value)

	v, err := db.HScan(key, nil, 100, true, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 4 {
		t.Fatal("invalid count", len(v))
	}

	v, err = db.HScan(key, []byte("19"), 1, false, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid count", len(v))
	} else if string(v[0].Field) != "222" {
		t.Fatal(string(v[0].Field))
	}

	v, err = db.HRevScan(key, []byte("19"), 1, false, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid count", len(v))
	} else if string(v[0].Field) != "1234" {
		t.Fatal(string(v[0].Field))
	}

}

func TestDBSScan(t *testing.T) {
	db := getTestDB()
	key := []byte("scan_s_key")

	db.SAdd(key, []byte("1"), []byte("222"), []byte("19"), []byte("1234"))

	v, err := db.SScan(key, nil, 100, true, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 4 {
		t.Fatal("invalid count", len(v))
	}

	v, err = db.SScan(key, []byte("19"), 1, false, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid count", len(v))
	} else if string(v[0]) != "222" {
		t.Fatal(string(v[0]))
	}

	v, err = db.SRevScan(key, []byte("19"), 1, false, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid count", len(v))
	} else if string(v[0]) != "1234" {
		t.Fatal(string(v[0]))
	}

}

func TestDBZScan(t *testing.T) {
	db := getTestDB()
	key := []byte("scan_z_key")

	db.ZAdd(key, ScorePair{1, []byte("1")}, ScorePair{2, []byte("222")}, ScorePair{3, []byte("19")}, ScorePair{4, []byte("1234")})

	v, err := db.ZScan(key, nil, 100, true, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 4 {
		t.Fatal("invalid count", len(v))
	}

	v, err = db.ZScan(key, []byte("19"), 1, false, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid count", len(v))
	} else if string(v[0].Member) != "222" {
		t.Fatal(string(v[0].Member))
	}

	v, err = db.ZRevScan(key, []byte("19"), 1, false, "")
	if err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal("invalid count", len(v))
	} else if string(v[0].Member) != "1234" {
		t.Fatal(string(v[0].Member))
	}

}
