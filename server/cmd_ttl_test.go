package server

import (
	"github.com/siddontang/ledisdb/client/go/ledis"
	"testing"
	"time"
)

func now() int64 {
	return time.Now().Unix()
}

func TestKVExpire(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	k := "a_ttl"
	c.Do("set", k, "123")

	//	expire + ttl
	exp := int64(10)
	if n, err := ledis.Int(c.Do("expire", k, exp)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("ttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != exp {
		t.Fatal(ttl)
	}

	//	expireat + ttl
	tm := now() + 3
	if n, err := ledis.Int(c.Do("expireat", k, tm)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("ttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != 3 {
		t.Fatal(ttl)
	}

	kErr := "not_exist_ttl"

	//	err - expire, expireat
	if n, err := ledis.Int(c.Do("expire", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("expireat", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("ttl", kErr)); err != nil || n != -1 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("persist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("expire", k, 10)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("persist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

}

func TestSetExpire(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	k := "set_ttl"
	c.Do("sadd", k, "123")

	//	expire + ttl
	exp := int64(10)
	if n, err := ledis.Int(c.Do("sexpire", k, exp)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("sttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != exp {
		t.Fatal(ttl)
	}

	//	expireat + ttl
	tm := now() + 3
	if n, err := ledis.Int(c.Do("sexpireat", k, tm)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("sttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != 3 {
		t.Fatal(ttl)
	}

	kErr := "not_exist_ttl"

	//	err - expire, expireat
	if n, err := ledis.Int(c.Do("sexpire", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("sexpireat", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("sttl", kErr)); err != nil || n != -1 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("spersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("sexpire", k, 10)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("spersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

}

func TestListExpire(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	k := "list_ttl"
	c.Do("rpush", k, "123")

	//	expire + ttl
	exp := int64(10)
	if n, err := ledis.Int(c.Do("lexpire", k, exp)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("lttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != exp {
		t.Fatal(ttl)
	}

	//	expireat + ttl
	tm := now() + 3
	if n, err := ledis.Int(c.Do("lexpireat", k, tm)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("lttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != 3 {
		t.Fatal(ttl)
	}

	kErr := "not_exist_ttl"

	//	err - expire, expireat
	if n, err := ledis.Int(c.Do("lexpire", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("lexpireat", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("lttl", kErr)); err != nil || n != -1 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("lpersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("lexpire", k, 10)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("lpersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

}

func TestHashExpire(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	k := "hash_ttl"
	c.Do("hset", k, "f", 123)

	//	expire + ttl
	exp := int64(10)
	if n, err := ledis.Int(c.Do("hexpire", k, exp)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("httl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != exp {
		t.Fatal(ttl)
	}

	//	expireat + ttl
	tm := now() + 3
	if n, err := ledis.Int(c.Do("hexpireat", k, tm)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("httl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != 3 {
		t.Fatal(ttl)
	}

	kErr := "not_exist_ttl"

	//	err - expire, expireat
	if n, err := ledis.Int(c.Do("hexpire", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("hexpireat", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("httl", kErr)); err != nil || n != -1 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("hpersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hexpire", k, 10)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hpersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

}

func TestZsetExpire(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	k := "zset_ttl"
	c.Do("zadd", k, 123, "m")

	//	expire + ttl
	exp := int64(10)
	if n, err := ledis.Int(c.Do("zexpire", k, exp)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("zttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != exp {
		t.Fatal(ttl)
	}

	//	expireat + ttl
	tm := now() + 3
	if n, err := ledis.Int(c.Do("zexpireat", k, tm)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("zttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != 3 {
		t.Fatal(ttl)
	}

	kErr := "not_exist_ttl"

	//	err - expire, expireat
	if n, err := ledis.Int(c.Do("zexpire", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("zexpireat", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("zttl", kErr)); err != nil || n != -1 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("zpersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zexpire", k, 10)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zpersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

}

func TestBitmapExpire(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	k := "bit_ttl"
	c.Do("bsetbit", k, 0, 1)

	//	expire + ttl
	exp := int64(10)
	if n, err := ledis.Int(c.Do("bexpire", k, exp)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("bttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != exp {
		t.Fatal(ttl)
	}

	//	expireat + ttl
	tm := now() + 3
	if n, err := ledis.Int(c.Do("bexpireat", k, tm)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := ledis.Int64(c.Do("bttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != 3 {
		t.Fatal(ttl)
	}

	kErr := "not_exist_ttl"

	//	err - expire, expireat
	if n, err := ledis.Int(c.Do("bexpire", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("bexpireat", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("bttl", kErr)); err != nil || n != -1 {
		t.Fatal(false)
	}

	if n, err := ledis.Int(c.Do("bpersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("bexpire", k, 10)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("bpersist", k)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

}
