package server

import (
	"os"
	"sync"
	"testing"

	"github.com/siddontang/goredis"
	"github.com/siddontang/ledisdb/config"
)

var testAppOnce sync.Once
var testAppAuthOnce sync.Once
var testApp *App

var testLedisClient *goredis.Client
var testLedisClientAuth *goredis.Client

func getTestConnAuth(password string) *goredis.PoolConn {
	startTestAppAuth(password)
	conn, _ := testLedisClientAuth.Get()
	return conn
}

func newTestLedisClientAuth() {
	testLedisClientAuth = goredis.NewClient("127.0.0.1:20000", "")
	testLedisClientAuth.SetMaxIdleConns(4)
}

func startTestAppAuth(password string) {
	f := func() {
		newTestLedisClientAuth()

		cfg := config.NewConfigDefault()
		cfg.DataDir = "/tmp/testdb_auth"
		os.RemoveAll(cfg.DataDir)

		cfg.Addr = "127.0.0.1:20000"
		cfg.HttpAddr = "127.0.0.1:20001"
		cfg.AuthPassword = password

		os.RemoveAll(cfg.DataDir)

		var err error
		testApp, err = NewApp(cfg)
		if err != nil {
			println(err.Error())
			panic(err)
		}

		go testApp.Run()
	}

	testAppAuthOnce.Do(f)
}

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

		os.RemoveAll(cfg.DataDir)

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
