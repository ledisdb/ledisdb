package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"github.com/siddontang/ledisdb/config"
	"os"
	"testing"
	"time"
)

func TestDumpRestore(t *testing.T) {
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

	testDumpRestore(c, "dump", "mtest_a", t)
	testDumpRestore(c, "ldump", "mtest_la", t)
	testDumpRestore(c, "hdump", "mtest_ha", t)
	testDumpRestore(c, "sdump", "mtest_sa", t)
	testDumpRestore(c, "zdump", "mtest_za", t)
}

func testDumpRestore(c *ledis.Conn, dump string, key string, t *testing.T) {
	if data, err := ledis.Bytes(c.Do(dump, key)); err != nil {
		t.Fatal(err)
	} else if _, err := c.Do("restore", key, 0, data); err != nil {
		t.Fatal(err)
	}
}

func TestMigrate(t *testing.T) {
	data_dir := "/tmp/test_migrate"
	os.RemoveAll(data_dir)

	s1Cfg := config.NewConfigDefault()
	s1Cfg.DataDir = fmt.Sprintf("%s/s1", data_dir)
	s1Cfg.Addr = "127.0.0.1:11185"

	s2Cfg := config.NewConfigDefault()
	s2Cfg.DataDir = fmt.Sprintf("%s/s2", data_dir)
	s2Cfg.Addr = "127.0.0.1:11186"

	s1, err := NewApp(s1Cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s1.Close()

	s2, err := NewApp(s2Cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	go s1.Run()

	go s2.Run()

	time.Sleep(1 * time.Second)

	c1 := ledis.NewConn(s1Cfg.Addr)
	defer c1.Close()

	c2 := ledis.NewConn(s2Cfg.Addr)
	defer c2.Close()

	if _, err = c1.Do("set", "a", "1"); err != nil {
		t.Fatal(err)
	}

	timeout := 30000
	if _, err = c1.Do("xmigrate", "127.0.0.1", 11186, "KV", "a", 0, timeout); err != nil {
		t.Fatal(err)
	}

	if s, err := ledis.String(c2.Do("get", "a")); err != nil {
		t.Fatal(err)
	} else if s != "1" {
		t.Fatal(s, "must 1")
	}

	if s, err := ledis.String(c1.Do("get", "a")); err != nil && err != ledis.ErrNil {
		t.Fatal(err)
	} else if s != "" {
		t.Fatal(s, "must empty")
	}

	if num, err := ledis.Int(c2.Do("xmigratedb", "127.0.0.1", 11185, "KV", 10, 0, timeout)); err != nil {
		t.Fatal(err)
	} else if num != 1 {
		t.Fatal(num, "must number 1")
	}

	if s, err := ledis.String(c1.Do("get", "a")); err != nil {
		t.Fatal(err)
	} else if s != "1" {
		t.Fatal(s, "must 1")
	}

	if s, err := ledis.String(c2.Do("get", "a")); err != nil && err != ledis.ErrNil {
		t.Fatal(err)
	} else if s != "" {
		t.Fatal(s, "must empty")
	}

}
