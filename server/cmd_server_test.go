package server

import (
	"github.com/siddontang/goredis"
	"testing"
)

func TestXSelect(t *testing.T) {
	c1 := getTestConn()
	defer c1.Close()

	c2 := getTestConn()
	defer c2.Close()

	_, err := c1.Do("XSELECT", "1", "THEN", "SET", "tmp_select_key", "1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = goredis.Int(c2.Do("GET", "tmp_select_key"))
	if err != goredis.ErrNil {
		t.Fatal(err)
	}

	n, _ := goredis.Int(c2.Do("XSELECT", "1", "THEN", "GET", "tmp_select_key"))
	if n != 1 {
		t.Fatal(n)
	}

	n, _ = goredis.Int(c2.Do("GET", "tmp_select_key"))
	if n != 1 {
		t.Fatal(n)
	}

	c1.Do("SELECT", 0)
	c2.Do("SELECT", 0)

}
