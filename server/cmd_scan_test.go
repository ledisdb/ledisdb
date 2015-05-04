package server

import (
	"fmt"
	"os"
	"testing"

	"github.com/siddontang/goredis"
	"github.com/siddontang/ledisdb/config"
)

func TestScan(t *testing.T) {
	cfg := config.NewConfigDefault()
	cfg.DataDir = "/tmp/test_scan"
	cfg.Addr = "127.0.0.1:11185"

	os.RemoveAll(cfg.DataDir)

	s, err := NewApp(cfg)
	if err != nil {
		t.Fatal(err)
	}
	go s.Run()
	defer s.Close()

	c := goredis.NewClient(cfg.Addr, "")
	c.SetMaxIdleConns(1)
	defer c.Close()

	testKVScan(t, c)
	testHashKeyScan(t, c)
	testListKeyScan(t, c)
	testZSetKeyScan(t, c)
	testSetKeyScan(t, c)

}

func checkScanValues(t *testing.T, ay interface{}, values ...interface{}) {
	a, err := goredis.Strings(ay, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(a) != len(values) {
		t.Fatal(fmt.Sprintf("len %d != %d", len(a), len(values)))
	}

	for i, v := range a {
		if string(v) != fmt.Sprintf("%v", values[i]) {
			t.Fatal(fmt.Sprintf("%d %s != %v", string(v), values[i]))
		}
	}
}

func checkScan(t *testing.T, c *goredis.Client, tp string) {
	if ay, err := goredis.Values(c.Do("XSCAN", tp, "", "count", 5)); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "4" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 0, 1, 2, 3, 4)
	}

	if ay, err := goredis.Values(c.Do("XSCAN", tp, "4", "count", 6)); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 5, 6, 7, 8, 9)
	}

}

func testKVScan(t *testing.T, c *goredis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("set", fmt.Sprintf("%d", i), []byte("value")); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "KV")
}

func testHashKeyScan(t *testing.T, c *goredis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("hset", fmt.Sprintf("%d", i), fmt.Sprintf("%d", i), []byte("value")); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "HASH")
}

func testListKeyScan(t *testing.T, c *goredis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("lpush", fmt.Sprintf("%d", i), fmt.Sprintf("%d", i)); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "LIST")
}

func testZSetKeyScan(t *testing.T, c *goredis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("zadd", fmt.Sprintf("%d", i), i, []byte("value")); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "ZSET")
}

func testSetKeyScan(t *testing.T, c *goredis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("sadd", fmt.Sprintf("%d", i), fmt.Sprintf("%d", i)); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "SET")
}

func TestHashScan(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := "scan_hash"
	c.Do("HMSET", key, "a", 1, "b", 2)

	if ay, err := goredis.Values(c.Do("XHSCAN", key, "")); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else {
		checkScanValues(t, ay[1], "a", 1, "b", 2)
	}
}

func TestSetScan(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := "scan_set"
	c.Do("SADD", key, "a", "b")

	if ay, err := goredis.Values(c.Do("XSSCAN", key, "")); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else {
		checkScanValues(t, ay[1], "a", "b")
	}

}

func TestZSetScan(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := "scan_zset"
	c.Do("ZADD", key, 1, "a", 2, "b")

	if ay, err := goredis.Values(c.Do("XZSCAN", key, "")); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else {
		checkScanValues(t, ay[1], "a", 1, "b", 2)
	}

}
