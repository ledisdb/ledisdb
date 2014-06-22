package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"strconv"
	"testing"
)

func TestHash(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("a")
	if n, err := ledis.Int(c.Do("hset", key, 1, 0)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hexists", key, 1)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hexists", key, -1)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hget", key, 1)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hset", key, 1, 1)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hget", key, 1)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hlen", key)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}
}

func testHashArray(ay []interface{}, checkValues ...int) error {
	if len(ay) != len(checkValues) {
		return fmt.Errorf("invalid return number %d != %d", len(ay), len(checkValues))
	}

	for i := 0; i < len(ay); i++ {
		if ay[i] == nil && checkValues[i] != 0 {
			return fmt.Errorf("must nil")
		} else if ay[i] != nil {
			v, ok := ay[i].([]byte)
			if !ok {
				return fmt.Errorf("invalid return data %d %v :%T", i, ay[i], ay[i])
			}

			d, _ := strconv.Atoi(string(v))

			if d != checkValues[i] {
				return fmt.Errorf("invalid data %d %s != %d", i, v, checkValues[i])
			}
		}
	}
	return nil
}

func TestHashM(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("b")
	if ok, err := ledis.String(c.Do("hmset", key, 1, 1, 2, 2, 3, 3)); err != nil {
		t.Fatal(err)
	} else if ok != OK {
		t.Fatal(ok)
	}

	if n, err := ledis.Int(c.Do("hlen", key)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if v, err := ledis.MultiBulk(c.Do("hmget", key, 1, 2, 3, 4)); err != nil {
		t.Fatal(err)
	} else {
		if err := testHashArray(v, 1, 2, 3, 0); err != nil {
			t.Fatal(err)
		}
	}

	if n, err := ledis.Int(c.Do("hdel", key, 1, 2, 3, 4)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hlen", key)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if v, err := ledis.MultiBulk(c.Do("hmget", key, 1, 2, 3, 4)); err != nil {
		t.Fatal(err)
	} else {
		if err := testHashArray(v, 0, 0, 0, 0); err != nil {
			t.Fatal(err)
		}
	}

	if n, err := ledis.Int(c.Do("hlen", key)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}
}

func TestHashIncr(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("c")
	if n, err := ledis.Int(c.Do("hincrby", key, 1, 1)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(err)
	}

	if n, err := ledis.Int(c.Do("hlen", key)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hincrby", key, 1, 10)); err != nil {
		t.Fatal(err)
	} else if n != 11 {
		t.Fatal(err)
	}

	if n, err := ledis.Int(c.Do("hlen", key)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hincrby", key, 1, -11)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(err)
	}

}

func TestHashGetAll(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("d")

	if ok, err := ledis.String(c.Do("hmset", key, 1, 1, 2, 2, 3, 3)); err != nil {
		t.Fatal(err)
	} else if ok != OK {
		t.Fatal(ok)
	}

	if v, err := ledis.MultiBulk(c.Do("hgetall", key)); err != nil {
		t.Fatal(err)
	} else {
		if err := testHashArray(v, 1, 1, 2, 2, 3, 3); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("hkeys", key)); err != nil {
		t.Fatal(err)
	} else {
		if err := testHashArray(v, 1, 2, 3); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("hvals", key)); err != nil {
		t.Fatal(err)
	} else {
		if err := testHashArray(v, 1, 2, 3); err != nil {
			t.Fatal(err)
		}
	}

	if n, err := ledis.Int(c.Do("hclear", key)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("hlen", key)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}
}
