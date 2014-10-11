package server

import (
	"github.com/siddontang/ledisdb/client/go/ledis"
	"github.com/siddontang/ledisdb/config"
	"os"
	"sync"
	"testing"
)

var testAppOnce sync.Once
var testApp *App

var testLedisClient *ledis.Client

func newTestLedisClient() {
	cfg := new(ledis.Config)
	cfg.Addr = "127.0.0.1:16380"
	cfg.MaxIdleConns = 4
	testLedisClient = ledis.NewClient(cfg)
}

func getTestConn() *ledis.Conn {
	startTestApp()
	return testLedisClient.Get()
}

func startTestApp() {
	f := func() {
		newTestLedisClient()

		cfg := config.NewConfigDefault()
		cfg.DataDir = "/tmp/testdb"
		os.RemoveAll(cfg.DataDir)

		cfg.Addr = "127.0.0.1:16380"

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
