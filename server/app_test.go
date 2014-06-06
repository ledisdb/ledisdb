package server

import (
	"github.com/garyburd/redigo/redis"
	"os"
	"sync"
	"testing"
)

var testAppOnce sync.Once
var testApp *App

var testPool *redis.Pool

func newTestRedisPool() {
	f := func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", "127.0.0.1:16380")
		if err != nil {
			return nil, err
		}

		return c, nil
	}

	testPool = redis.NewPool(f, 4)
}

func getTestConn() redis.Conn {
	startTestApp()
	return testPool.Get()
}

func startTestApp() {
	f := func() {
		newTestRedisPool()

		os.RemoveAll("/tmp/testdb")

		var d = []byte(`
            {
                "data_dir" : "/tmp/testdb",
                "addr" : "127.0.0.1:16380",
                "db" : {
                    "data_db" : {
                        "compression":true,
                        "block_size" : 32768,
                        "write_buffer_size" : 2097152,
                        "cache_size" : 20971520
                    }
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
