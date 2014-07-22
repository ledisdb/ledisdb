package server

import (
	"github.com/siddontang/ledisdb/client/go/ledis"
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

		os.RemoveAll("/tmp/testdb")

		var d = []byte(`
            {
                "data_dir" : "/tmp/testdb",
                "addr" : "127.0.0.1:16380",
                "db" : {        
                    "compression":true,
                    "block_size" : 32768,
                    "write_buffer_size" : 2097152,
                    "cache_size" : 20971520,
                    "max_open_files" : 1024
                }    
            }
            `)

		cfg, err := NewConfig(d)
		if err != nil {
			println(err.Error())
			panic(err)
		}

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
