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

var KB = config.KB
var MB = config.MB
var GB = config.GB

var name = flag.String("db_name", "goleveldb", "db name")
var number = flag.Int("n", 10000, "request number")
var clients = flag.Int("c", 50, "number of clients")
var round = flag.Int("r", 1, "benchmark round number")
var valueSize = flag.Int("vsize", 100, "kv value size")
var wg sync.WaitGroup

var db *store.DB

var loop int = 0

func bench(cmd string, f func()) {
	wg.Add(*clients)

	t1 := time.Now()
	for i := 0; i < *clients; i++ {
		go func() {
			for i := 0; i < loop; i++ {
				f()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	t2 := time.Now()

	fmt.Printf("%s: %0.2f op/s, %0.2fmb/s\n", cmd, (float64(*number) / t2.Sub(t1).Seconds()),
		float64(*valueSize*(*number))/(1024.0*1024.0*(t2.Sub(t1).Seconds())))
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

func setRocksDB(cfg *config.RocksDBConfig) {
	cfg.BlockSize = 64 * KB
	cfg.WriteBufferSize = 64 * MB
	cfg.MaxWriteBufferNum = 2
	cfg.MaxBytesForLevelBase = 512 * MB
	cfg.TargetFileSizeBase = 64 * MB
	cfg.BackgroundThreads = 4
	cfg.HighPriorityBackgroundThreads = 1
	cfg.MaxBackgroundCompactions = 3
	cfg.MaxBackgroundFlushes = 1
	cfg.CacheSize = 512 * MB
	cfg.EnableStatistics = true
	cfg.StatsDumpPeriodSec = 5
	cfg.Level0FileNumCompactionTrigger = 8
	cfg.MaxBytesForLevelMultiplier = 8
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	cfg := config.NewConfigDefault()
	cfg.DBPath = "./var/store_test"
	cfg.DBName = *name
	os.RemoveAll(cfg.DBPath)

	cfg.LevelDB.BlockSize = 32 * KB
	cfg.LevelDB.CacheSize = 512 * MB
	cfg.LevelDB.WriteBufferSize = 64 * MB
	cfg.LevelDB.MaxOpenFiles = 1000

	setRocksDB(&cfg.RocksDB)

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
