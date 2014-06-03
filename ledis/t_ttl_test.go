package ledis

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var m sync.Mutex

func TestKvExpire(t *testing.T) {
	db := getTestDB()
	m.Lock()
	defer m.Unlock()

	k := []byte("ttl_a")
	ek := []byte("ttl_b")
	db.Set(k, []byte("1"))

	if ok, _ := db.Expire(k, 10); ok != 1 {
		t.Fatal(ok)
	}

	//	err - expire on an inexisting key
	if ok, _ := db.Expire(ek, 10); ok != 0 {
		t.Fatal(ok)
	}

	//	err - duration is zero
	if ok, err := db.Expire(k, 0); err == nil || ok != 0 {
		t.Fatal(fmt.Sprintf("res = %d, err = %s", ok, err))
	}

	//	err - duration is negative
	if ok, err := db.Expire(k, -10); err == nil || ok != 0 {
		t.Fatal(fmt.Sprintf("res = %d, err = %s", ok, err))
	}
}

func TestKvExpireAt(t *testing.T) {
	db := getTestDB()
	m.Lock()
	defer m.Unlock()

	k := []byte("ttl_a")
	ek := []byte("ttl_b")
	db.Set(k, []byte("1"))

	now := time.Now().Unix()

	if ok, _ := db.ExpireAt(k, now+5); ok != 1 {
		t.Fatal(ok)
	}

	//	err - expire on an inexisting key
	if ok, _ := db.ExpireAt(ek, now+5); ok != 0 {
		t.Fatal(ok)
	}

	//	err - expire with the current time
	if ok, err := db.ExpireAt(k, now); err == nil || ok != 0 {
		t.Fatal(fmt.Sprintf("res = %d, err = %s", ok, err))
	}

	//	err - expire with the time before
	if ok, err := db.ExpireAt(k, now-5); err == nil || ok != 0 {
		t.Fatal(fmt.Sprintf("res = %d, err = %s", ok, err))
	}
}

func TestKvTtl(t *testing.T) {
	db := getTestDB()
	m.Lock()
	defer m.Unlock()

	k := []byte("ttl_a")
	ek := []byte("ttl_b")

	db.Set(k, []byte("1"))
	db.Expire(k, 2)

	if tRemain, _ := db.Ttl(k); tRemain != 2 {
		t.Fatal(tRemain)
	}

	//	err - check ttl on an inexisting key
	if tRemain, _ := db.Ttl(ek); tRemain != -1 {
		t.Fatal(tRemain)
	}

	db.Del(k)
	if tRemain, _ := db.Ttl(k); tRemain != -1 {
		t.Fatal(tRemain)
	}
}

func TestKvExpCompose(t *testing.T) {
	db := getTestDB()
	m.Lock()
	defer m.Unlock()

	k0 := []byte("ttl_a")
	k1 := []byte("ttl_b")
	k2 := []byte("ttl_c")

	db.Set(k0, k0)
	db.Set(k1, k1)
	db.Set(k2, k2)

	db.Expire(k0, 5)
	db.Expire(k1, 2)
	db.Expire(k2, 60)

	if tRemain, _ := db.Ttl(k0); tRemain != 5 {
		t.Fatal(tRemain)
	}
	if tRemain, _ := db.Ttl(k1); tRemain != 2 {
		t.Fatal(tRemain)
	}
	if tRemain, _ := db.Ttl(k2); tRemain != 60 {
		t.Fatal(tRemain)
	}

	// after 1 sec
	time.Sleep(1 * time.Second)
	if tRemain, _ := db.Ttl(k0); tRemain != 4 {
		t.Fatal(tRemain)
	}
	if tRemain, _ := db.Ttl(k1); tRemain != 1 {
		t.Fatal(tRemain)
	}

	// after 2 sec
	time.Sleep(2 * time.Second)
	if tRemain, _ := db.Ttl(k1); tRemain != -1 {
		t.Fatal(tRemain)
	}
	if v, _ := db.Get(k1); v != nil {
		t.Fatal(v)
	}

	if tRemain, _ := db.Ttl(k0); tRemain != 2 {
		t.Fatal(tRemain)
	}
	if v, _ := db.Get(k0); v == nil {
		t.Fatal(v)
	}

	// refresh the expiration of key
	if tRemain, _ := db.Ttl(k2); !(0 < tRemain && tRemain < 60) {
		t.Fatal(tRemain)
	}

	if ok, _ := db.Expire(k2, 100); ok != 1 {
		t.Fatal(false)
	}

	if tRemain, _ := db.Ttl(k2); tRemain != 100 {
		t.Fatal(tRemain)
	}

	//	expire an inexisting key
	if ok, _ := db.Expire(k1, 10); ok == 1 {
		t.Fatal(false)
	}
	if tRemain, _ := db.Ttl(k1); tRemain != -1 {
		t.Fatal(tRemain)
	}

	return
}
