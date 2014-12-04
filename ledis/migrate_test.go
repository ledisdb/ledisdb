package ledis

import (
	"github.com/siddontang/ledisdb/config"
	"os"
	"testing"
)

func TestMigrate(t *testing.T) {
	cfg1 := config.NewConfigDefault()
	cfg1.DataDir = "/tmp/test_ledisdb_migrate1"
	os.RemoveAll(cfg1.DataDir)

	defer os.RemoveAll(cfg1.DataDir)

	l1, _ := Open(cfg1)
	defer l1.Close()

	cfg2 := config.NewConfigDefault()
	cfg2.DataDir = "/tmp/test_ledisdb_migrate2"
	os.RemoveAll(cfg2.DataDir)

	defer os.RemoveAll(cfg2.DataDir)

	l2, _ := Open(cfg2)
	defer l2.Close()

	db1, _ := l1.Select(0)
	db2, _ := l2.Select(0)

	key := []byte("a")
	lkey := []byte("a")
	hkey := []byte("a")
	skey := []byte("a")
	zkey := []byte("a")
	value := []byte("1")

	db1.Set(key, value)

	if data, err := db1.Dump(key); err != nil {
		t.Fatal(err)
	} else if err := db2.Restore(key, 0, data); err != nil {
		t.Fatal(err)
	}

	db1.RPush(lkey, []byte("1"), []byte("2"), []byte("3"))

	if data, err := db1.LDump(lkey); err != nil {
		t.Fatal(err)
	} else if err := db2.Restore(lkey, 0, data); err != nil {
		t.Fatal(err)
	}

	db1.SAdd(skey, []byte("1"), []byte("2"), []byte("3"))

	if data, err := db1.SDump(skey); err != nil {
		t.Fatal(err)
	} else if err := db2.Restore(skey, 0, data); err != nil {
		t.Fatal(err)
	}

	db1.HMset(hkey, FVPair{[]byte("a"), []byte("1")}, FVPair{[]byte("b"), []byte("2")}, FVPair{[]byte("c"), []byte("3")})

	if data, err := db1.HDump(hkey); err != nil {
		t.Fatal(err)
	} else if err := db2.Restore(hkey, 0, data); err != nil {
		t.Fatal(err)
	}

	db1.ZAdd(zkey, ScorePair{1, []byte("a")}, ScorePair{2, []byte("b")}, ScorePair{3, []byte("c")})

	if data, err := db1.ZDump(zkey); err != nil {
		t.Fatal(err)
	} else if err := db2.Restore(zkey, 0, data); err != nil {
		t.Fatal(err)
	}

	if err := checkLedisEqual(l1, l2); err != nil {
		t.Fatal(err)
	}
}
