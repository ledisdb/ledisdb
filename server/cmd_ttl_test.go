package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"testing"
	"time"
)

func now() int64 {
	return time.Now().Unix()
}

func TestExpire(t *testing.T) {
	// test for kv, list, hash, set, zset, bitmap in all
	ttlType := []string{"k", "l", "h", "s", "z", "b"}

	var (
		expire   string
		expireat string
		ttl      string
		persist  string
		key      string
	)

	c := getTestConn()
	defer c.Close()

	idx := 1
	for _, tt := range ttlType {
		if tt == "k" {
			expire = "expire"
			expireat = "expireat"
			ttl = "ttl"
			persist = "persist"

		} else {
			expire = fmt.Sprintf("%sexpire", tt)
			expireat = fmt.Sprintf("%sexpireat", tt)
			ttl = fmt.Sprintf("%sttl", tt)
			persist = fmt.Sprintf("%spersist", tt)
		}

		switch tt {
		case "k":
			key = "kv_ttl"
			c.Do("set", key, "123")
		case "l":
			key = "list_ttl"
			c.Do("rpush", key, "123")
		case "h":
			key = "hash_ttl"
			c.Do("hset", key, "a", "123")
		case "s":
			key = "set_ttl"
			c.Do("sadd", key, "123")
		case "z":
			key = "zset_ttl"
			c.Do("zadd", key, 123, "a")
		case "b":
			key = "bitmap_ttl"
			c.Do("bsetbit", key, 0, 1)
		}

		//	expire + ttl
		exp := int64(10)
		if n, err := ledis.Int(c.Do(expire, key, exp)); err != nil {
			t.Fatal(err)
		} else if n != 1 {
			t.Fatal(n)
		}

		if ttl, err := ledis.Int64(c.Do(ttl, key)); err != nil {
			t.Fatal(err)
		} else if ttl == -1 {
			t.Fatal("no ttl")
		}

		//	expireat + ttl
		tm := now() + 3
		if n, err := ledis.Int(c.Do(expireat, key, tm)); err != nil {
			t.Fatal(err)
		} else if n != 1 {
			t.Fatal(n)
		}

		if ttl, err := ledis.Int64(c.Do(ttl, key)); err != nil {
			t.Fatal(err)
		} else if ttl == -1 {
			t.Fatal("no ttl")
		}

		kErr := "not_exist_ttl"

		//	err - expire, expireat
		if n, err := ledis.Int(c.Do(expire, kErr, tm)); err != nil || n != 0 {
			t.Fatal(false)
		}

		if n, err := ledis.Int(c.Do(expireat, kErr, tm)); err != nil || n != 0 {
			t.Fatal(false)
		}

		if n, err := ledis.Int(c.Do(ttl, kErr)); err != nil || n != -1 {
			t.Fatal(false)
		}

		if n, err := ledis.Int(c.Do(persist, key)); err != nil {
			t.Fatal(err)
		} else if n != 1 {
			t.Fatal(n)
		}

		if n, err := ledis.Int(c.Do(expire, key, 10)); err != nil {
			t.Fatal(err)
		} else if n != 1 {
			t.Fatal(n)
		}

		if n, err := ledis.Int(c.Do(persist, key)); err != nil {
			t.Fatal(err)
		} else if n != 1 {
			t.Fatal(n)
		}

		idx++
	}

}
