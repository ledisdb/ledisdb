package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"strconv"
	"testing"
)

func testListIndex(key []byte, index int64, v int) error {
	c := getTestConn()
	defer c.Close()

	n, err := ledis.Int(c.Do("lindex", key, index))
	if err == ledis.ErrNil && v != 0 {
		return fmt.Errorf("must nil")
	} else if err != nil && err != ledis.ErrNil {
		return err
	} else if n != v {
		return fmt.Errorf("index err number %d != %d", n, v)
	}

	return nil
}

func testListRange(key []byte, start int64, stop int64, checkValues ...int) error {
	c := getTestConn()
	defer c.Close()

	vs, err := ledis.MultiBulk(c.Do("lrange", key, start, stop))
	if err != nil {
		return err
	}

	if len(vs) != len(checkValues) {
		return fmt.Errorf("invalid return number %d != %d", len(vs), len(checkValues))
	}

	var n int
	for i, v := range vs {
		if d, ok := v.([]byte); ok {
			n, err = strconv.Atoi(string(d))
			if err != nil {
				return err
			} else if n != checkValues[i] {
				return fmt.Errorf("invalid data %d: %d != %d", i, n, checkValues[i])
			}
		} else {
			return fmt.Errorf("invalid data %v %T", v, v)
		}
	}

	return nil
}

func TestList(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("a")

	if n, err := ledis.Int(c.Do("lpush", key, 1)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("rpush", key, 2)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("rpush", key, 3)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("llen", key)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	//for ledis-cli a 1 2 3
	// 127.0.0.1:6379> lrange a 0 0
	// 1) "1"
	if err := testListRange(key, 0, 0, 1); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a 0 1
	// 1) "1"
	// 2) "2"

	if err := testListRange(key, 0, 1, 1, 2); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a 0 5
	// 1) "1"
	// 2) "2"
	// 3) "3"
	if err := testListRange(key, 0, 5, 1, 2, 3); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a -1 5
	// 1) "3"
	if err := testListRange(key, -1, 5, 3); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a -5 -1
	// 1) "1"
	// 2) "2"
	// 3) "3"
	if err := testListRange(key, -5, -1, 1, 2, 3); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a -2 -1
	// 1) "2"
	// 2) "3"
	if err := testListRange(key, -2, -1, 2, 3); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a -1 -2
	// (empty list or set)
	if err := testListRange(key, -1, -2); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a -1 2
	// 1) "3"
	if err := testListRange(key, -1, 2, 3); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a -5 5
	// 1) "1"
	// 2) "2"
	// 3) "3"
	if err := testListRange(key, -5, 5, 1, 2, 3); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a -1 0
	// (empty list or set)
	if err := testListRange(key, -1, 0); err != nil {
		t.Fatal(err)
	}

	if err := testListRange([]byte("empty list"), 0, 100); err != nil {
		t.Fatal(err)
	}

	// 127.0.0.1:6379> lrange a -1 -1
	// 1) "3"
	if err := testListRange(key, -1, -1, 3); err != nil {
		t.Fatal(err)
	}

	if err := testListIndex(key, -1, 3); err != nil {
		t.Fatal(err)
	}

	if err := testListIndex(key, 0, 1); err != nil {
		t.Fatal(err)
	}

	if err := testListIndex(key, 1, 2); err != nil {
		t.Fatal(err)
	}

	if err := testListIndex(key, 2, 3); err != nil {
		t.Fatal(err)
	}

	if err := testListIndex(key, 5, 0); err != nil {
		t.Fatal(err)
	}

	if err := testListIndex(key, -1, 3); err != nil {
		t.Fatal(err)
	}

	if err := testListIndex(key, -2, 2); err != nil {
		t.Fatal(err)
	}

	if err := testListIndex(key, -3, 1); err != nil {
		t.Fatal(err)
	}
}

func TestListMPush(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("b")
	if n, err := ledis.Int(c.Do("rpush", key, 1, 2, 3)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if err := testListRange(key, 0, 3, 1, 2, 3); err != nil {
		t.Fatal(err)
	}

	if n, err := ledis.Int(c.Do("lpush", key, 1, 2, 3)); err != nil {
		t.Fatal(err)
	} else if n != 6 {
		t.Fatal(n)
	}

	if err := testListRange(key, 0, 6, 3, 2, 1, 1, 2, 3); err != nil {
		t.Fatal(err)
	}
}

func TestPop(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("c")
	if n, err := ledis.Int(c.Do("rpush", key, 1, 2, 3, 4, 5, 6)); err != nil {
		t.Fatal(err)
	} else if n != 6 {
		t.Fatal(n)
	}

	if v, err := ledis.Int(c.Do("lpop", key)); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Fatal(v)
	}

	if v, err := ledis.Int(c.Do("rpop", key)); err != nil {
		t.Fatal(err)
	} else if v != 6 {
		t.Fatal(v)
	}

	if n, err := ledis.Int(c.Do("lpush", key, 1)); err != nil {
		t.Fatal(err)
	} else if n != 5 {
		t.Fatal(n)
	}

	if err := testListRange(key, 0, 5, 1, 2, 3, 4, 5); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 5; i++ {
		if v, err := ledis.Int(c.Do("lpop", key)); err != nil {
			t.Fatal(err)
		} else if v != i {
			t.Fatal(v)
		}
	}

	if n, err := ledis.Int(c.Do("llen", key)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	c.Do("rpush", key, 1, 2, 3, 4, 5)

	if n, err := ledis.Int(c.Do("lclear", key)); err != nil {
		t.Fatal(err)
	} else if n != 5 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("llen", key)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

}

func TestListErrorParams(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if _, err := c.Do("lpush", "test_lpush"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("rpush", "test_rpush"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lpop", "test_lpop", "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("rpop", "test_rpop", "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("llen", "test_llen", "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lindex", "test_lindex"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lrange", "test_lrange"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lclear"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lmclear"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lexpire"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lexpireat"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lttl"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("lpersist"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

}
