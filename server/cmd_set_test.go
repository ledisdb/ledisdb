package server

import (
	"github.com/siddontang/ledisdb/client/go/ledis"
	"testing"
)

func TestSet(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key1 := "testdb_cmd_set_1"
	key2 := "testdb_cmd_set_2"

	if n, err := ledis.Int(c.Do("sadd", key1, 0, 1)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("sadd", key2, 0, 1, 2, 3)); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("scard", key1)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.MultiBulk(c.Do("sdiff", key2, key1)); err != nil {
		t.Fatal(err)
	} else if len(n) != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("sdiffstore", []byte("cmd_set_em1"), key2, key1)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.MultiBulk(c.Do("sunion", key1, key2)); err != nil {
		t.Fatal(err)
	} else if len(n) != 4 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("sunionstore", []byte("cmd_set_em2"), key1, key2)); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if n, err := ledis.MultiBulk(c.Do("sinter", key1, key2)); err != nil {
		t.Fatal(err)
	} else if len(n) != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("sinterstore", []byte("cmd_set_em3"), key1, key2)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("srem", key1, 0, 1)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("sismember", key2, 0)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.MultiBulk(c.Do("smembers", key2)); err != nil {
		t.Fatal(err)
	} else if len(n) != 4 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("sclear", key2)); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	c.Do("sadd", key1, 0)
	c.Do("sadd", key2, 1)
	if n, err := ledis.Int(c.Do("smclear", key1, key2)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

}

func TestSetErrorParams(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if _, err := c.Do("sadd", "test_sadd"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("scard"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("scard", "k1", "k2"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sdiff"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sdiffstore", "dstkey"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sinter"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sinterstore", "dstkey"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sunion"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sunionstore", "dstkey"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sismember", "k1"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sismember", "k1", "m1", "m2"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("smembers"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("smembers", "k1", "k2"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("srem"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("srem", "key"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sclear"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sclear", "k1", "k2"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("smclear"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sexpire", "set_expire"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sexpire", "set_expire", "aaa"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sexpireat", "set_expireat"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sexpireat", "set_expireat", "aaa"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("sttl"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("spersist"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

}
