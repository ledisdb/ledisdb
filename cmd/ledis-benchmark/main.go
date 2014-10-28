package main

import (
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var ip = flag.String("ip", "127.0.0.1", "redis/ledis/ssdb server ip")
var port = flag.Int("port", 6380, "redis/ledis/ssdb server port")
var number = flag.Int("n", 1000, "request number")
var clients = flag.Int("c", 50, "number of clients")
var round = flag.Int("r", 1, "benchmark round number")
var valueSize = flag.Int("vsize", 100, "kv value size")
var tests = flag.String("t", "", "only run the comma separated list of tests, set,get,del,lpush,lrange,lpop,hset,hget,hdel,zadd,zincr,zrange,zrevrange,zdel")
var wg sync.WaitGroup

var client *ledis.Client

var loop int = 0

func waitBench(cmd string, args ...interface{}) {
	c := client.Get()
	defer c.Close()

	_, err := c.Do(strings.ToUpper(cmd), args...)
	if err != nil {
		fmt.Printf("do %s error %s\n", cmd, err.Error())
		return
	}
}

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

	fmt.Printf("%s: %s %0.3f micros/op, %0.2fop/s\n",
		cmd,
		d.String(),
		float64(d.Nanoseconds()/1e3)/float64(*number),
		float64(*number)/d.Seconds())
}

var kvSetBase int64 = 0
var kvGetBase int64 = 0
var kvIncrBase int64 = 0
var kvDelBase int64 = 0

func benchSet() {
	f := func() {
		value := make([]byte, *valueSize)
		n := atomic.AddInt64(&kvSetBase, 1)
		waitBench("set", n, value)
	}

	bench("set", f)
}

func benchGet() {
	f := func() {
		n := atomic.AddInt64(&kvGetBase, 1)
		waitBench("get", n)
	}

	bench("get", f)
}

func benchRandGet() {
	f := func() {
		n := rand.Int() % *number
		waitBench("get", n)
	}

	bench("randget", f)
}

func benchDel() {
	f := func() {
		n := atomic.AddInt64(&kvDelBase, 1)
		waitBench("del", n)
	}

	bench("del", f)
}

func benchPushList() {
	f := func() {
		value := make([]byte, 100)
		waitBench("rpush", "mytestlist", value)
	}

	bench("rpush", f)
}

func benchRangeList10() {
	f := func() {
		waitBench("lrange", "mytestlist", 0, 10)
	}

	bench("lrange10", f)
}

func benchRangeList50() {
	f := func() {
		waitBench("lrange", "mytestlist", 0, 50)
	}

	bench("lrange50", f)
}

func benchRangeList100() {
	f := func() {
		waitBench("lrange", "mytestlist", 0, 100)
	}

	bench("lrange100", f)
}

func benchPopList() {
	f := func() {
		waitBench("lpop", "mytestlist")
	}

	bench("lpop", f)
}

var hashSetBase int64 = 0
var hashIncrBase int64 = 0
var hashGetBase int64 = 0
var hashDelBase int64 = 0

func benchHset() {
	f := func() {
		value := make([]byte, 100)

		n := atomic.AddInt64(&hashSetBase, 1)
		waitBench("hset", "myhashkey", n, value)
	}

	bench("hset", f)
}

func benchHGet() {
	f := func() {
		n := atomic.AddInt64(&hashGetBase, 1)
		waitBench("hget", "myhashkey", n)
	}

	bench("hget", f)
}

func benchHRandGet() {
	f := func() {
		n := rand.Int() % *number
		waitBench("hget", "myhashkey", n)
	}

	bench("hrandget", f)
}

func benchHDel() {
	f := func() {
		n := atomic.AddInt64(&hashDelBase, 1)
		waitBench("hdel", "myhashkey", n)
	}

	bench("hdel", f)
}

var zsetAddBase int64 = 0
var zsetDelBase int64 = 0
var zsetIncrBase int64 = 0

func benchZAdd() {
	f := func() {
		member := make([]byte, 16)
		n := atomic.AddInt64(&zsetAddBase, 1)
		waitBench("zadd", "myzsetkey", n, member)
	}

	bench("zadd", f)
}

func benchZDel() {
	f := func() {
		n := atomic.AddInt64(&zsetDelBase, 1)
		waitBench("zrem", "myzsetkey", n)
	}

	bench("zrem", f)
}

func benchZIncr() {
	f := func() {
		n := atomic.AddInt64(&zsetIncrBase, 1)
		waitBench("zincrby", "myzsetkey", 1, n)
	}

	bench("zincrby", f)
}

func benchZRangeByScore() {
	f := func() {
		waitBench("zrangebyscore", "myzsetkey", 0, rand.Int(), "withscores", "limit", rand.Int()%100, 100)
	}

	bench("zrangebyscore", f)
}

func benchZRangeByRank() {
	f := func() {
		waitBench("zrange", "myzsetkey", 0, rand.Int()%100)
	}

	bench("zrange", f)
}

func benchZRevRangeByScore() {
	f := func() {
		waitBench("zrevrangebyscore", "myzsetkey", 0, rand.Int(), "withscores", "limit", rand.Int()%100, 100)
	}

	bench("zrevrangebyscore", f)
}

func benchZRevRangeByRank() {
	f := func() {
		waitBench("zrevrange", "myzsetkey", 0, rand.Int()%100)
	}

	bench("zrevrange", f)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	if *number <= 0 {
		panic("invalid number")
		return
	}

	if *clients <= 0 || *number < *clients {
		panic("invalid client number")
		return
	}

	loop = *number / *clients

	addr := fmt.Sprintf("%s:%d", *ip, *port)

	cfg := new(ledis.Config)
	cfg.Addr = addr
	cfg.MaxIdleConns = *clients
	cfg.ReadBufferSize = 10240
	cfg.WriteBufferSize = 10240
	client = ledis.NewClient(cfg)

	if *round <= 0 {
		*round = 1
	}

	runAll := true
	ts := strings.Split(*tests, ",")
	if len(ts) > 0 && len(ts[0]) != 0 {
		runAll = false
	}

	needTest := make(map[string]struct{})
	for _, s := range ts {
		needTest[strings.ToLower(s)] = struct{}{}
	}

	checkTest := func(cmd string) bool {
		if runAll {
			return true
		} else if _, ok := needTest[cmd]; ok {
			return ok
		}
		return false
	}

	for i := 0; i < *round; i++ {
		if checkTest("set") {
			benchSet()
		}

		if checkTest("get") {
			benchGet()
			benchRandGet()
		}

		if checkTest("del") {
			benchDel()
		}

		if checkTest("lpush") {
			benchPushList()
		}

		if checkTest("lrange") {
			benchRangeList10()
			benchRangeList50()
			benchRangeList100()
		}

		if checkTest("lpop") {
			benchPopList()
		}

		if checkTest("hset") {
			benchHset()
		}

		if checkTest("hget") {
			benchHGet()
			benchHRandGet()
		}

		if checkTest("hdel") {
			benchHDel()
		}

		if checkTest("zadd") {
			benchZAdd()
		}

		if checkTest("zincr") {
			benchZIncr()
		}

		if checkTest("zrange") {
			benchZRangeByRank()
			benchZRangeByScore()
		}

		if checkTest("zrevrange") {
			//rev is too slow in leveldb, rocksdb or other
			//maybe disable for huge data benchmark
			benchZRevRangeByRank()
			benchZRevRangeByScore()
		}

		if checkTest("zdel") {
			benchZDel()
		}

		println("")
	}
}
