package main

import (
	"flag"
	"fmt"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
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

var ldb *ledis.Ledis
var db *ledis.DB

var loop int = 0

func bench(cmd string, f func()) {
	wg.Add(*clients)

	t1 := time.Now()
	for i := 0; i < *clients; i++ {
		go func() {
			for j := 0; j < loop; j++ {
				f()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	t2 := time.Now()

	d := t2.Sub(t1)
	fmt.Printf("%s %s: %0.3f micros/op, %0.2fmb/s %0.2fop/s\n",
		cmd,
		d.String(),
		float64(d.Nanoseconds()/1e3)/float64(*number),
		float64((*valueSize+16)*(*number))/(1024.0*1024.0*(d.Seconds())),
		float64(*number)/d.Seconds())
}

var kvSetBase int64 = 0
var kvGetBase int64 = 0

var value []byte

func benchSet() {
	f := func() {
		n := atomic.AddInt64(&kvSetBase, 1)

		db.Set(num.Int64ToBytes(n), value)
	}

	bench("set", f)
}

func benchGet() {
	kvGetBase = 0
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

var kvGetSliceBase int64 = 0

func benchGetSlice() {
	kvGetSliceBase = 0
	f := func() {
		n := atomic.AddInt64(&kvGetSliceBase, 1)
		v, err := db.GetSlice(num.Int64ToBytes(n))
		if err != nil {
			println(err.Error())
		} else if v != nil {
			v.Free()
		}
	}

	bench("getslice", f)
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

	value = make([]byte, *valueSize)

	cfg := config.NewConfigDefault()
	cfg.DataDir = "./var/ledis_dbbench"
	cfg.DBName = *name
	os.RemoveAll(cfg.DBPath)
	defer os.RemoveAll(cfg.DBPath)

	os.MkdirAll(cfg.DBPath, 0755)

	cfg.LevelDB.BlockSize = 32 * KB
	cfg.LevelDB.CacheSize = 512 * MB
	cfg.LevelDB.WriteBufferSize = 64 * MB
	cfg.LevelDB.MaxOpenFiles = 1000

	setRocksDB(&cfg.RocksDB)

	var err error
	ldb, err = ledis.Open(cfg)
	if err != nil {
		println(err.Error())
		return
	}

	db, _ = ldb.Select(0)

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
		benchGetSlice()
		benchGet()
		benchGetSlice()

		println("")
	}
}
