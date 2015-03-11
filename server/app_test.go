package server

import (
	"github.com/siddontang/goredis"
	"github.com/siddontang/ledisdb/config"
	"os"
	"sync"
	"testing"
)

var testAppOnce sync.Once
var testApp *App

var testLedisClient *goredis.Client

func newTestLedisClient() {
	testLedisClient = goredis.NewClient("127.0.0.1:16380", "")
	testLedisClient.SetMaxIdleConns(4)
}

func getTestConn() *goredis.PoolConn {
	startTestApp()
	conn, _ := testLedisClient.Get()
	return conn
}

func startTestApp() {
	f := func() {
		newTestLedisClient()

		cfg := config.NewConfigDefault()
		cfg.DataDir = "/tmp/testdb"
		os.RemoveAll(cfg.DataDir)

		cfg.Addr = "127.0.0.1:16380"
		cfg.HttpAddr = "127.0.0.1:21181"

		os.RemoveAll("/tmp/testdb")

		var err error
		testApp, err = NewApp(cfg)
		if err != nil {
			println(err.Error())
			panic(err)
		}

		go testApp.Run()
	}

	testAppOnce.Do(f)
}

func TestApp(t *testing.T) {
	startTestApp()
}
