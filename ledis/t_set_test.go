package ledis

import (
	"fmt"
	"testing"
	"time"
)

func TestSetCodec(t *testing.T) {
	db := getTestDB()

	key := []byte("key")
	member := []byte("member")

	ek := db.sEncodeSizeKey(key)
	if k, err := db.sDecodeSizeKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	}

	ek = db.sEncodeSetKey(key, member)
	if k, m, err := db.sDecodeSetKey(ek); err != nil {
		t.Fatal(err)
	} else if string(k) != "key" {
		t.Fatal(string(k))
	} else if string(m) != "member" {
		t.Fatal(string(m))
	}
}

func TestDBSet(t *testing.T) {
	db := getTestDB()

	key := []byte("testdb_set_a")
	member := []byte("member")
	key1 := []byte("testdb_set_a1")
	key2 := []byte("testdb_set_a2")
	member1 := []byte("testdb_set_m1")
	member2 := []byte("testdb_set_m2")

	// if n, err := db.sSetItem(key, []byte("m1")); err != nil {
	// 	t.Fatal(err)
	// } else if n != 1 {
	// 	t.Fatal(n)
	// }

	// if size, err := db.sIncrSize(key, 1); err != nil {
	// 	t.Fatal(err)
	// } else if size != 1 {
	// 	t.Fatal(size)
	// }

	if n, err := db.SAdd(key, member); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if cnt, err := db.SCard(key); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Fatal(cnt)
	}

	if n, err := db.SIsMember(key, member); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if v, err := db.SMembers(key); err != nil {
		t.Fatal(err)
	} else if string(v[0]) != "member" {
		t.Fatal(string(v[0]))
	}

	if n, err := db.SRem(key, member); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	db.SAdd(key1, member1, member2)

	if n, err := db.SClear(key1); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	db.SAdd(key1, member1, member2)
	db.SAdd(key2, member1, member2, []byte("xxx"))

	if n, _ := db.SCard(key2); n != 3 {
		t.Fatal(n)
	}
	if n, err := db.SMclear(key1, key2); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	db.SAdd(key2, member1, member2)
	if n, err := db.SExpire(key2, 3600); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := db.SExpireAt(key2, time.Now().Unix()+3600); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := db.STTL(key2); err != nil {
		t.Fatal(err)
	} else if n < 0 {
		t.Fatal(n)
	}

	if n, err := db.SPersist(key2); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

}

func TestSetOperation(t *testing.T) {
	db := getTestDB()
	testUnion(db, t)
	testInter(db, t)
	testDiff(db, t)

}

func testUnion(db *DB, t *testing.T) {

	key := []byte("testdb_set_union_1")
	key1 := []byte("testdb_set_union_2")
	key2 := []byte("testdb_set_union_2")
	// member1 := []byte("testdb_set_m1")
	// member2 := []byte("testdb_set_m2")

	m1 := []byte("m1")
	m2 := []byte("m2")
	m3 := []byte("m3")
	db.SAdd(key, m1, m2)
	db.SAdd(key1, m1, m2, m3)
	db.SAdd(key2, m2, m3)
	if _, err := db.sUnionGeneric(key, key2); err != nil {
		t.Fatal(err)
	}

	if _, err := db.SUnion(key, key2); err != nil {
		t.Fatal(err)
	}

	dstkey := []byte("union_dsk")
	db.SAdd(dstkey, []byte("x"))
	if num, err := db.SUnionStore(dstkey, key1, key2); err != nil {
		t.Fatal(err)
	} else if num != 3 {
		t.Fatal(num)
	}

	if _, err := db.SMembers(dstkey); err != nil {
		t.Fatal(err)
	}

	if n, err := db.SCard(dstkey); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	v1, _ := db.SUnion(key1, key2)
	v2, _ := db.SUnion(key2, key1)
	if len(v1) != len(v2) {
		t.Fatal(v1, v2)
	}

	v1, _ = db.SUnion(key, key1, key2)
	v2, _ = db.SUnion(key, key2, key1)
	if len(v1) != len(v2) {
		t.Fatal(v1, v2)
	}

	if v, err := db.SUnion(key, key); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(v)
	}

	empKey := []byte("0")
	if v, err := db.SUnion(key, empKey); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(v)
	}
}

func testInter(db *DB, t *testing.T) {
	key1 := []byte("testdb_set_inter_1")
	key2 := []byte("testdb_set_inter_2")
	key3 := []byte("testdb_set_inter_3")

	m1 := []byte("m1")
	m2 := []byte("m2")
	m3 := []byte("m3")
	m4 := []byte("m4")

	db.SAdd(key1, m1, m2)
	db.SAdd(key2, m2, m3, m4)
	db.SAdd(key3, m2, m4)

	if v, err := db.sInterGeneric(key1, key2); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal(v)
	}

	if v, err := db.SInter(key1, key2); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal(v)
	}

	dstKey := []byte("inter_dsk")
	if n, err := db.SInterStore(dstKey, key1, key2); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	k1 := []byte("set_k1")
	k2 := []byte("set_k2")

	db.SAdd(k1, m1, m3, m4)
	db.SAdd(k2, m2, m3)
	if n, err := db.SInterStore([]byte("set_xxx"), k1, k2); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	v1, _ := db.SInter(key1, key2)
	v2, _ := db.SInter(key2, key1)
	if len(v1) != len(v2) {
		t.Fatal(v1, v2)
	}

	v1, _ = db.SInter(key1, key2, key3)
	v2, _ = db.SInter(key2, key3, key1)
	if len(v1) != len(v2) {
		t.Fatal(v1, v2)
	}

	if v, err := db.SInter(key1, key1); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(v)
	}

	empKey := []byte("0")
	if v, err := db.SInter(key1, empKey); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(v)
	}

	if v, err := db.SInter(empKey, key2); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(v)
	}
}

func testDiff(db *DB, t *testing.T) {
	key0 := []byte("testdb_set_diff_0")
	key1 := []byte("testdb_set_diff_1")
	key2 := []byte("testdb_set_diff_2")
	key3 := []byte("testdb_set_diff_3")

	m1 := []byte("m1")
	m2 := []byte("m2")
	m3 := []byte("m3")
	m4 := []byte("m4")

	db.SAdd(key1, m1, m2)
	db.SAdd(key2, m2, m3, m4)
	db.SAdd(key3, m3)

	if _, err := db.sDiffGeneric(key1, key2); err != nil {
		t.Fatal(err)
	}

	if v, err := db.SDiff(key1, key2); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal(v)
	}

	dstKey := []byte("diff_dsk")
	if n, err := db.SDiffStore(dstKey, key1, key2); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if v, err := db.SDiff(key2, key1); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(v)
	}

	if v, err := db.SDiff(key1, key2, key3); err != nil {
		t.Fatal(err)
	} else if len(v) != 1 {
		t.Fatal(v) //return 1
	}

	if v, err := db.SDiff(key2, key2); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(v)
	}

	if v, err := db.SDiff(key0, key1); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(v)
	}

	if v, err := db.SDiff(key1, key0); err != nil {
		t.Fatal(err)
	} else if len(v) != 2 {
		t.Fatal(v)
	}
}

func TestSFlush(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("%d", i)
		if _, err := db.SAdd([]byte(key), []byte("v")); err != nil {
			t.Fatal(err.Error())
		}
	}

	if v, err := db.SScan(nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 2000 {
		t.Fatal("invalid value ", len(v))
	}

	if n, err := db.sFlush(); err != nil {
		t.Fatal(err.Error())
	} else if n != 2000 {
		t.Fatal("invalid value ", n)
	}

	if v, err := db.SScan(nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 0 {
		t.Fatal("invalid value length ", len(v))
	}

}
