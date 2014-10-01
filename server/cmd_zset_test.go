package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"reflect"
	"strconv"
	"testing"
)

func TestZSet(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("myzset")
	if n, err := ledis.Int(c.Do("zadd", key, 3, "a", 4, "b")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(n)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zadd", key, 1, "a", 2, "b")); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(n)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zadd", key, 3, "c", 4, "d")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if s, err := ledis.Int(c.Do("zscore", key, "c")); err != nil {
		t.Fatal(err)
	} else if s != 3 {
		t.Fatal(s)
	}

	if n, err := ledis.Int(c.Do("zrem", key, "d", "e")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zincrby", key, 4, "c")); err != nil {
		t.Fatal(err)
	} else if n != 7 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zincrby", key, -4, "c")); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zincrby", key, 4, "d")); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zrem", key, "a", "b", "c", "d")); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

}

func TestZSetCount(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("myzset")
	if _, err := ledis.Int(c.Do("zadd", key, 1, "a", 2, "b", 3, "c", 4, "d")); err != nil {
		t.Fatal(err)
	}

	if n, err := ledis.Int(c.Do("zcount", key, 2, 4)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcount", key, 4, 4)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcount", key, 4, 3)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcount", key, "(2", 4)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcount", key, "2", "(4")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcount", key, "(2", "(4")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcount", key, "-inf", "+inf")); err != nil {
		t.Fatal(err)
	} else if n != 4 {
		t.Fatal(n)
	}

	c.Do("zadd", key, 3, "e")

	if n, err := ledis.Int(c.Do("zcount", key, "(2", "(4")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	c.Do("zrem", key, "a", "b", "c", "d", "e")
}

func TestZSetRank(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("myzset")
	if _, err := ledis.Int(c.Do("zadd", key, 1, "a", 2, "b", 3, "c", 4, "d")); err != nil {
		t.Fatal(err)
	}

	if n, err := ledis.Int(c.Do("zrank", key, "c")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if _, err := ledis.Int(c.Do("zrank", key, "e")); err != ledis.ErrNil {
		t.Fatal(err)
	}

	if n, err := ledis.Int(c.Do("zrevrank", key, "c")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if _, err := ledis.Int(c.Do("zrevrank", key, "e")); err != ledis.ErrNil {
		t.Fatal(err)
	}
}

func testZSetRange(ay []interface{}, checkValues ...interface{}) error {
	if len(ay) != len(checkValues) {
		return fmt.Errorf("invalid return number %d != %d", len(ay), len(checkValues))
	}

	for i := 0; i < len(ay); i++ {
		v, ok := ay[i].([]byte)
		if !ok {
			return fmt.Errorf("invalid data %d %v %T", i, ay[i], ay[i])
		}

		switch cv := checkValues[i].(type) {
		case string:
			if string(v) != cv {
				return fmt.Errorf("not equal %s != %s", v, checkValues[i])
			}
		default:
			if s, _ := strconv.Atoi(string(v)); s != checkValues[i] {
				return fmt.Errorf("not equal %s != %v", v, checkValues[i])
			}
		}

	}

	return nil
}

func TestZSetRangeScore(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("myzset_range")
	if _, err := ledis.Int(c.Do("zadd", key, 1, "a", 2, "b", 3, "c", 4, "d")); err != nil {
		t.Fatal(err)
	}

	if v, err := ledis.MultiBulk(c.Do("zrangebyscore", key, 1, 4, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "a", 1, "b", 2, "c", 3, "d", 4); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrangebyscore", key, 1, 4, "withscores", "limit", 1, 2)); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "b", 2, "c", 3); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrangebyscore", key, "-inf", "+inf", "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "a", 1, "b", 2, "c", 3, "d", 4); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrangebyscore", key, "(1", "(4")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "b", "c"); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrevrangebyscore", key, 4, 1, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "d", 4, "c", 3, "b", 2, "a", 1); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrevrangebyscore", key, 4, 1, "withscores", "limit", 1, 2)); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "c", 3, "b", 2); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrevrangebyscore", key, "+inf", "-inf", "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "d", 4, "c", 3, "b", 2, "a", 1); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrevrangebyscore", key, "(4", "(1")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "c", "b"); err != nil {
			t.Fatal(err)
		}
	}

	if n, err := ledis.Int(c.Do("zremrangebyscore", key, 2, 3)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if v, err := ledis.MultiBulk(c.Do("zrangebyscore", key, 1, 4)); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "a", "d"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestZSetRange(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("myzset_range_rank")
	if _, err := ledis.Int(c.Do("zadd", key, 1, "a", 2, "b", 3, "c", 4, "d")); err != nil {
		t.Fatal(err)
	}

	if v, err := ledis.MultiBulk(c.Do("zrange", key, 0, 3, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "a", 1, "b", 2, "c", 3, "d", 4); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrange", key, 1, 4, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "b", 2, "c", 3, "d", 4); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrange", key, -2, -1, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "c", 3, "d", 4); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrange", key, 0, -1, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "a", 1, "b", 2, "c", 3, "d", 4); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrange", key, -1, -2, "withscores")); err != nil {
		t.Fatal(err)
	} else if len(v) != 0 {
		t.Fatal(len(v))
	}

	if v, err := ledis.MultiBulk(c.Do("zrevrange", key, 0, 4, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "d", 4, "c", 3, "b", 2, "a", 1); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrevrange", key, 0, -1, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "d", 4, "c", 3, "b", 2, "a", 1); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrevrange", key, 2, 3, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "b", 2, "a", 1); err != nil {
			t.Fatal(err)
		}
	}

	if v, err := ledis.MultiBulk(c.Do("zrevrange", key, -2, -1, "withscores")); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "b", 2, "a", 1); err != nil {
			t.Fatal(err)
		}
	}

	if n, err := ledis.Int(c.Do("zremrangebyrank", key, 2, 3)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if v, err := ledis.MultiBulk(c.Do("zrange", key, 0, 4)); err != nil {
		t.Fatal(err)
	} else {
		if err := testZSetRange(v, "a", "b"); err != nil {
			t.Fatal(err)
		}
	}

	if n, err := ledis.Int(c.Do("zclear", key)); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int(c.Do("zcard", key)); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

}

func TestZsetErrorParams(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	//zadd
	if _, err := c.Do("zadd", "test_zadd"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zadd", "test_zadd", "a", "b", "c"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zadd", "test_zadd", "-a", "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zadd", "test_zad", "0.1", "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zcard
	if _, err := c.Do("zcard"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zscore
	if _, err := c.Do("zscore", "test_zscore"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zrem
	if _, err := c.Do("zrem", "test_zrem"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zincrby
	if _, err := c.Do("zincrby", "test_zincrby"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zincrby", "test_zincrby", 0.1, "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zcount
	if _, err := c.Do("zcount", "test_zcount"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zcount", "test_zcount", "-inf", "=inf"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zcount", "test_zcount", 0.1, 0.1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zrank
	if _, err := c.Do("zrank", "test_zrank"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zrevzrank
	if _, err := c.Do("zrevrank", "test_zrevrank"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zremrangebyrank
	if _, err := c.Do("zremrangebyrank", "test_zremrangebyrank"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zremrangebyrank", "test_zremrangebyrank", 0.1, 0.1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zremrangebyscore
	if _, err := c.Do("zremrangebyscore", "test_zremrangebyscore"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zremrangebyscore", "test_zremrangebyscore", "-inf", "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zremrangebyscore", "test_zremrangebyscore", 0, "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zrange
	if _, err := c.Do("zrange", "test_zrange"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zrange", "test_zrange", 0, 1, "withscore"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zrange", "test_zrange", 0, 1, "withscores", "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zrevrange, almost same as zrange
	if _, err := c.Do("zrevrange", "test_zrevrange"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zrangebyscore
	if _, err := c.Do("zrangebyscore", "test_zrangebyscore"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zrangebyscore", "test_zrangebyscore", 0, 1, "withscore"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zrangebyscore", "test_zrangebyscore", 0, 1, "withscores", "limit"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zrangebyscore", "test_zrangebyscore", 0, 1, "withscores", "limi", 1, 1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zrangebyscore", "test_zrangebyscore", 0, 1, "withscores", "limit", "a", 1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("zrangebyscore", "test_zrangebyscore", 0, 1, "withscores", "limit", 1, "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zrevrangebyscore, almost same as zrangebyscore
	if _, err := c.Do("zrevrangebyscore", "test_zrevrangebyscore"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zclear
	if _, err := c.Do("zclear"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zmclear
	if _, err := c.Do("zmclear"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zexpire
	if _, err := c.Do("zexpire", "test_zexpire"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zexpireat
	if _, err := c.Do("zexpireat", "test_zexpireat"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zttl
	if _, err := c.Do("zttl"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//zpersist
	if _, err := c.Do("zpersist"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

}

func TestZUnionStore(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if _, err := c.Do("zadd", "k1", "1", "one"); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := c.Do("zadd", "k1", "2", "two"); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := c.Do("zadd", "k2", "1", "two"); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := c.Do("zadd", "k2", "2", "three"); err != nil {
		t.Fatal(err.Error())
	}

	if n, err := ledis.Int64(c.Do("zunionstore", "out", "2", "k1", "k2", "weights", "1", "2")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 3 {
			t.Fatal("invalid value ", n)
		}
	}

	if n, err := ledis.Int64(c.Do("zunionstore", "out", "2", "k1", "k2", "weights", "1", "2", "aggregate", "min")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 3 {
			t.Fatal("invalid value ", n)
		}
	}

	if n, err := ledis.Int64(c.Do("zunionstore", "out", "2", "k1", "k2", "aggregate", "max")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 3 {
			t.Fatal("invalid value ", n)
		}
	}

	if n, err := ledis.Int64(c.Do("zscore", "out", "two")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 2 {
			t.Fatal("invalid value ", n)
		}
	}
}

func TestZInterStore(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if _, err := c.Do("zadd", "k1", "1", "one"); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := c.Do("zadd", "k1", "2", "two"); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := c.Do("zadd", "k2", "1", "two"); err != nil {
		t.Fatal(err.Error())
	}

	if _, err := c.Do("zadd", "k2", "2", "three"); err != nil {
		t.Fatal(err.Error())
	}

	if n, err := ledis.Int64(c.Do("zinterstore", "out", "2", "k1", "k2", "weights", "1", "2")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 1 {
			t.Fatal("invalid value ", n)
		}
	}

	if n, err := ledis.Int64(c.Do("zinterstore", "out", "2", "k1", "k2", "aggregate", "min", "weights", "1", "2")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 1 {
			t.Fatal("invalid value ", n)
		}
	}

	if n, err := ledis.Int64(c.Do("zinterstore", "out", "2", "k1", "k2", "aggregate", "sum")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 1 {
			t.Fatal("invalid value ", n)
		}
	}

	if n, err := ledis.Int64(c.Do("zscore", "out", "two")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 3 {
			t.Fatal("invalid value ", n)
		}
	}

	if _, err := c.Do("zadd", "k3", "3", "three"); err != nil {
		t.Fatal(err.Error())
	}

	if n, err := ledis.Int64(c.Do("zinterstore", "out", "3", "k1", "k2", "k3", "aggregate", "sum")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 0 {
			t.Fatal("invalid value ", n)
		}
	}

	if _, err := c.Do("zadd", "k3", "3", "two"); err != nil {
		t.Fatal(err.Error())
	}

	if n, err := ledis.Int64(c.Do("zinterstore", "out", "3", "k1", "k2", "k3", "aggregate", "sum", "weights", "3", "2", "2")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 1 {
			t.Fatal("invalid value ", n)
		}
	}

	if n, err := ledis.Int64(c.Do("zscore", "out", "two")); err != nil {
		t.Fatal(err.Error())
	} else {
		if n != 14 {
			t.Fatal("invalid value ", n)
		}
	}
}

func TestZSetLex(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("myzlexset")
	if _, err := c.Do("zadd", key,
		0, "a", 0, "b", 0, "c", 0, "d", 0, "e", 0, "f", 0, "g"); err != nil {
		t.Fatal(err)
	}

	if ay, err := ledis.Strings(c.Do("zrangebylex", key, "-", "[c")); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ay, []string{"a", "b", "c"}) {
		t.Fatal("must equal")
	}

	if ay, err := ledis.Strings(c.Do("zrangebylex", key, "-", "(c")); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ay, []string{"a", "b"}) {
		t.Fatal("must equal")
	}

	if ay, err := ledis.Strings(c.Do("zrangebylex", key, "[aaa", "(g")); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ay, []string{"b", "c", "d", "e", "f"}) {
		t.Fatal("must equal")
	}

	if n, err := ledis.Int64(c.Do("zlexcount", key, "-", "(c")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := ledis.Int64(c.Do("zremrangebylex", key, "[aaa", "(g")); err != nil {
		t.Fatal(err)
	} else if n != 5 {
		t.Fatal(n)
	}

	if n, err := ledis.Int64(c.Do("zlexcount", key, "-", "+")); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

}
