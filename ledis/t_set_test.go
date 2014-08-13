package ledis

import (
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

	// tx := db.setTx

	// if n := db.sDelete(tx, key1); n != 2 {
	// 	t.Fatal(n)
	// }

	if n, err := db.SClear(key1); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	db.SAdd(key2, member1, member2)
	db.SAdd(key1, member1, member2)

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
