package server

import (
	"github.com/garyburd/redigo/redis"
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
	if n, err := redis.Int(c.Do("expire", k, exp)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := redis.Int64(c.Do("ttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != exp {
		t.Fatal(ttl)
	}

	//	expireat + ttl
	tm := now() + 3
	if n, err := redis.Int(c.Do("expireat", k, tm)); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if ttl, err := redis.Int64(c.Do("ttl", k)); err != nil {
		t.Fatal(err)
	} else if ttl != 3 {
		t.Fatal(ttl)
	}

	kErr := "not_exist_ttl"

	//	err - expire, expireat
	if n, err := redis.Int(c.Do("expire", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := redis.Int(c.Do("expireat", kErr, tm)); err != nil || n != 0 {
		t.Fatal(false)
	}

	if n, err := redis.Int(c.Do("ttl", kErr)); err != nil || n != -1 {
		t.Fatal(false)
	}

}
