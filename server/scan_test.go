package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"github.com/siddontang/ledisdb/config"
	"os"
	"testing"
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

	cc := new(ledis.Config)
	cc.Addr = cfg.Addr
	cc.MaxIdleConns = 1
	c := ledis.NewClient(cc)
	defer c.Close()

	testKVScan(t, c)
	testHashScan(t, c)
	testListScan(t, c)
	testZSetScan(t, c)
	testSetScan(t, c)
	testBitScan(t, c)

}

func checkScanValues(t *testing.T, ay interface{}, values ...int) {
	a, err := ledis.Strings(ay, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(a) != len(values) {
		t.Fatal(fmt.Sprintf("len %d != %d", len(a), len(values)))
	}

	for i, v := range a {
		if string(v) != fmt.Sprintf("%d", values[i]) {
			t.Fatal(fmt.Sprintf("%d %s != %d", string(v), values[i]))
		}
	}
}

func checkScan(t *testing.T, c *ledis.Client, cmd string) {
	if ay, err := ledis.Values(c.Do(cmd, "", "count", 5)); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "4" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 0, 1, 2, 3, 4)
	}

	if ay, err := ledis.Values(c.Do(cmd, "4", "count", 6)); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 5, 6, 7, 8, 9)
	}

}

func checkRevScan(t *testing.T, c *ledis.Client, cmd string) {
	if ay, err := ledis.Values(c.Do(cmd, "", "count", 5)); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "5" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 9, 8, 7, 6, 5)
	}

	if ay, err := ledis.Values(c.Do(cmd, "5", "count", 6)); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 4, 3, 2, 1, 0)
	}

}

func testKVScan(t *testing.T, c *ledis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("set", fmt.Sprintf("%d", i), []byte("value")); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "xscan")
	checkRevScan(t, c, "xrevscan")
}

func testHashScan(t *testing.T, c *ledis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("hset", fmt.Sprintf("%d", i), fmt.Sprintf("%d", i), []byte("value")); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "hxscan")
	checkRevScan(t, c, "hxrevscan")
}

func testListScan(t *testing.T, c *ledis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("lpush", fmt.Sprintf("%d", i), fmt.Sprintf("%d", i)); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "lxscan")
	checkRevScan(t, c, "lxrevscan")
}

func testZSetScan(t *testing.T, c *ledis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("zadd", fmt.Sprintf("%d", i), i, []byte("value")); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "zxscan")
	checkRevScan(t, c, "zxrevscan")
}

func testSetScan(t *testing.T, c *ledis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("sadd", fmt.Sprintf("%d", i), fmt.Sprintf("%d", i)); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "sxscan")
	checkRevScan(t, c, "sxrevscan")
}

func testBitScan(t *testing.T, c *ledis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("bsetbit", fmt.Sprintf("%d", i), 1024, 1); err != nil {
			t.Fatal(err)
		}
	}

	checkScan(t, c, "bxscan")
	checkRevScan(t, c, "bxrevscan")
}
