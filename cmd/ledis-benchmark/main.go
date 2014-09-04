package main

import (
	crand "crypto/rand"
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var ip = flag.String("ip", "127.0.0.1", "redis/ledis/ssdb server ip")
var port = flag.Int("port", 6380, "redis/ledis/ssdb server port")
var number = flag.Int("n", 1000, "request number")
var clients = flag.Int("c", 50, "number of clients")
var reverse = flag.Bool("rev", false, "enable zset rev benchmark")
var round = flag.Int("r", 1, "benchmark round number")

var wg sync.WaitGroup

var client *ledis.Client

var loop int = 0

func waitBench(cmd string, args ...interface{}) {
	defer wg.Done()

	c := client.Get()
	defer c.Close()

	for i := 0; i < loop; i++ {
		_, err := c.Do(cmd, args...)
		if err != nil {
			fmt.Printf("do %s error %s", cmd, err.Error())
			return
		}
	}
}

func bench(cmd string, f func()) {
	wg.Add(*clients)

	t1 := time.Now().UnixNano()
	for i := 0; i < *clients; i++ {
		go f()
	}

	wg.Wait()

	t2 := time.Now().UnixNano()

	delta := float64(t2-t1) / float64(time.Second)

	fmt.Printf("%s: %0.2f requests per second\n", cmd, (float64(*number) / delta))
}

var kvSetBase int64 = 0
var kvGetBase int64 = 0
var kvIncrBase int64 = 0
var kvDelBase int64 = 0

func benchSet() {
	f := func() {
		value := make([]byte, 100)
		crand.Read(value)
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
		n := rand.Int()
		waitBench("get", n)
	}

	bench("get", f)
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
		value := make([]byte, 10)
		crand.Read(value)
		waitBench("rpush", "mytestlist", value)
	}

	bench("rpush", f)
}

func benchRangeList10() {
	f := func() {
		waitBench("lrange", "mytestlist", 0, 10)
	}

	bench("lrange", f)
}

func benchRangeList50() {
	f := func() {
		waitBench("lrange", "mytestlist", 0, 50)
	}

	bench("lrange", f)
}

func benchRangeList100() {
	f := func() {
		waitBench("lrange", "mytestlist", 0, 100)
	}

	bench("lrange", f)
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
		crand.Read(value)

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
		n := rand.Int()
		waitBench("hget", "myhashkey", n)
	}

	bench("hget", f)
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
		crand.Read(member)
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
	client = ledis.NewClient(cfg)

	if *round <= 0 {
		*round = 1
	}

	for i := 0; i < *round; i++ {
		benchSet()
		benchGet()
		benchRandGet()
		benchDel()

		benchPushList()
		benchRangeList10()
		benchRangeList50()
		benchRangeList100()
		benchPopList()

		benchHset()
		benchHGet()
		benchHRandGet()
		benchHDel()

		benchZAdd()
		benchZIncr()
		benchZRangeByRank()
		benchZRangeByScore()

		//rev is too slow in leveldb, rocksdb or other
		//maybe disable for huge data benchmark
		if *reverse == true {
			benchZRevRangeByRank()
			benchZRevRangeByScore()
		}

		benchZDel()

		println("")
	}
}
