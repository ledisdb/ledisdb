package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"github.com/siddontang/ledisdb/config"
	"os"
	"testing"
)

func TestScan(t *testing.T) {
	cfg := new(config.Config)
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

func testKVScan(t *testing.T, c *ledis.Client) {
	for i := 0; i < 10; i++ {
		if _, err := c.Do("set", fmt.Sprintf("%d", i), []byte("value")); err != nil {
			t.Fatal(err)
		}
	}

	if ay, err := ledis.Values(c.Do("scan", "", "count", 5)); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "4" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 0, 1, 2, 3, 4)
	}

	if ay, err := ledis.Values(c.Do("scan", "4", "count", 6)); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 5, 6, 7, 8, 9)
	}

	if ay, err := ledis.Values(c.Do("scan", "4", "count", 6, "inclusive")); err != nil {
		t.Fatal(err)
	} else if len(ay) != 2 {
		t.Fatal(len(ay))
	} else if n := ay[0].([]byte); string(n) != "9" {
		t.Fatal(string(n))
	} else {
		checkScanValues(t, ay[1], 4, 5, 6, 7, 8, 9)
	}

}
