package server

import (
	"github.com/siddontang/ledisdb/client/go/ledis"
	"testing"
)

func TestKV(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if ok, err := ledis.String(c.Do("set", "a", "1234")); err != nil {
		t.Fatal(err)
	} else if ok != OK {
		t.Fatal(ok)
	}

	if n, err := ledis.Int(c.Do("setnx", "a", "123")); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("setnx", "b", "123")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ok, err := ledis.String(c.Do("setex", "xx", 10, "hello world")); err != nil {
		t.Fatal(err)
	} else if ok != OK {
		t.Fatal(ok)
	}

	if v, err := ledis.String(c.Do("get", "a")); err != nil {
		t.Fatal(err)
	} else if v != "1234" {
		t.Fatal(v)
	}

	if v, err := ledis.String(c.Do("getset", "a", "123")); err != nil {
		t.Fatal(err)
	} else if v != "1234" {
		t.Fatal(v)
	}

	if v, err := ledis.String(c.Do("get", "a")); err != nil {
		t.Fatal(err)
	} else if v != "123" {
		t.Fatal(v)
	}

	if n, err := ledis.Int(c.Do("exists", "a")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("exists", "empty_key_test")); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if _, err := ledis.Int(c.Do("del", "a", "b")); err != nil {
		t.Fatal(err)
	}

	if n, err := ledis.Int(c.Do("exists", "a")); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("exists", "b")); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}
}

func TestKVM(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if ok, err := ledis.String(c.Do("mset", "a", "1", "b", "2")); err != nil {
		t.Fatal(err)
	} else if ok != OK {
		t.Fatal(ok)
	}

	if v, err := ledis.MultiBulk(c.Do("mget", "a", "b", "c")); err != nil {
		t.Fatal(err)
	} else if len(v) != 3 {
		t.Fatal(len(v))
	} else {
		if vv, ok := v[0].([]byte); !ok || string(vv) != "1" {
			t.Fatal("not 1")
		}

		if vv, ok := v[1].([]byte); !ok || string(vv) != "2" {
			t.Fatal("not 2")
		}

		if v[2] != nil {
			t.Fatal("must nil")
		}
	}
}

func TestKVIncrDecr(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if n, err := ledis.Int64(c.Do("incr", "n")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int64(c.Do("incr", "n")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int64(c.Do("decr", "n")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int64(c.Do("incrby", "n", 10)); err != nil {
		t.Fatal(err)
	} else if n != 11 {
		t.Fatal(n)
	}

	if n, err := ledis.Int64(c.Do("decrby", "n", 10)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}
}

func TestKVErrorParams(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if _, err := c.Do("get", "a", "b", "c"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

	if _, err := c.Do("set", "a", "b", "c"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

	if _, err := c.Do("getset", "a", "b", "c"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

	if _, err := c.Do("setnx", "a", "b", "c"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

	if _, err := c.Do("exists", "a", "b"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

	if _, err := c.Do("incr", "a", "b"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

	if _, err := c.Do("incrby", "a"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

	if _, err := c.Do("decrby", "a"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

	if _, err := c.Do("del"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("mset"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("mset", "a", "b", "c"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("mget"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("expire"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("expire", "a", "b"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("expireat"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("expireat", "a", "b"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("ttl"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("persist"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("setex", "a", "blah", "hello world"); err == nil {
		t.Fatalf("invalid err %v", err)
	}

}
