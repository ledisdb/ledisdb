package server

import (
	"github.com/siddontang/ledisdb/client/go/ledis"
	"testing"
)

func TestMigrate(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	var err error
	_, err = c.Do("set", "mtest_a", "1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Do("rpush", "mtest_la", "1", "2", "3")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Do("hmset", "mtest_ha", "a", "1", "b", "2")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Do("sadd", "mtest_sa", "1", "2", "3")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Do("zadd", "mtest_za", 1, "a", 2, "b", 3, "c")
	if err != nil {
		t.Fatal(err)
	}

	testMigrate(c, "dump", "mtest_a", t)
	testMigrate(c, "ldump", "mtest_la", t)
	testMigrate(c, "hdump", "mtest_ha", t)
	testMigrate(c, "sdump", "mtest_sa", t)
	testMigrate(c, "zdump", "mtest_za", t)
}

func testMigrate(c *ledis.Conn, dump string, key string, t *testing.T) {
	if data, err := ledis.Bytes(c.Do(dump, key)); err != nil {
		t.Fatal(err)
	} else if _, err := c.Do("restore", key, 0, data); err != nil {
		t.Fatal(err)
	}
}
