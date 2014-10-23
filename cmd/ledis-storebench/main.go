package main

import (
	crand "crypto/rand"
	"flag"
	"fmt"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/store"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var name = flag.String("db_name", "goleveldb", "db name")
var number = flag.Int("n", 1000, "request number")
var clients = flag.Int("c", 50, "number of clients")
var round = flag.Int("r", 1, "benchmark round number")
var valueSize = flag.Int("vsize", 100, "kv value size")
var wg sync.WaitGroup

var db *store.DB

var loop int = 0

func bench(cmd string, f func()) {
	wg.Add(*clients)

	t1 := time.Now().UnixNano()
	for i := 0; i < *clients; i++ {
		go func() {
			for i := 0; i < loop; i++ {
				f()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	t2 := time.Now().UnixNano()

	delta := float64(t2-t1) / float64(time.Second)

	fmt.Printf("%s: %0.2f requests per second\n", cmd, (float64(*number) / delta))
}

var kvSetBase int64 = 0
var kvGetBase int64 = 0

func benchSet() {
	f := func() {
		value := make([]byte, *valueSize)
		crand.Read(value)
		n := atomic.AddInt64(&kvSetBase, 1)

		db.Put(num.Int64ToBytes(n), value)
	}

	bench("set", f)
}

func benchGet() {
	f := func() {
		n := atomic.AddInt64(&kvGetBase, 1)
		v, err := db.Get(num.Int64ToBytes(n))
		if err != nil {
			println(err.Error())
		} else if len(v) != *valueSize {
			println(len(v), *valueSize)
		}
	}

	bench("get", f)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	cfg := config.NewConfigDefault()
	cfg.DBPath = "./store_test"
	os.RemoveAll(cfg.DBPath)
	defer os.RemoveAll(cfg.DBPath)

	cfg.LevelDB.BlockSize = 32 * 1024
	cfg.LevelDB.CacheSize = 512 * 1024 * 1024
	cfg.LevelDB.WriteBufferSize = 64 * 1024 * 1024

	var err error
	db, err = store.Open(cfg)
	if err != nil {
		panic(err)
		return
	}

	if *number <= 0 {
		panic("invalid number")
		return
	}

	if *clients <= 0 || *number < *clients {
		panic("invalid client number")
		return
	}

	loop = *number / *clients

	if *round <= 0 {
		*round = 1
	}

	for i := 0; i < *round; i++ {
		benchSet()
		benchGet()

		println("")
	}
}
