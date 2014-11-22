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
var tests = flag.String("t", "set,get,randget,del,lpush,lrange,lpop,hset,hget,hdel,zadd,zincr,zrange,zrevrange,zdel", "only run the comma separated list of tests")
var wg sync.WaitGroup

var client *ledis.Client
var loop int = 0

func waitBench(c *ledis.Conn, cmd string, args ...interface{}) {
	_, err := c.Do(strings.ToUpper(cmd), args...)
	if err != nil {
		fmt.Printf("do %s error %s\n", cmd, err.Error())
	}

}

func bench(cmd string, f func(c *ledis.Conn)) {
	wg.Add(*clients)

	t1 := time.Now()
	for i := 0; i < *clients; i++ {
		go func() {
			c := client.Get()
			for j := 0; j < loop; j++ {
				f(c)
			}
			c.Close()
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
	f := func(c *ledis.Conn) {
		value := make([]byte, *valueSize)
		n := atomic.AddInt64(&kvSetBase, 1)
		waitBench(c, "SET", n, value)
	}

	bench("set", f)
}

func benchGet() {
	f := func(c *ledis.Conn) {
		n := atomic.AddInt64(&kvGetBase, 1)
		waitBench(c, "GET", n)
	}

	bench("get", f)
}

func benchRandGet() {
	f := func(c *ledis.Conn) {
		n := rand.Int() % *number
		waitBench(c, "GET", n)
	}

	bench("randget", f)
}

func benchDel() {
	f := func(c *ledis.Conn) {
		n := atomic.AddInt64(&kvDelBase, 1)
		waitBench(c, "DEL", n)
	}

	bench("del", f)
}

func benchPushList() {
	f := func(c *ledis.Conn) {
		value := make([]byte, 100)
		waitBench(c, "RPUSH", "mytestlist", value)
	}

	bench("rpush", f)
}

func benchRangeList10() {
	f := func(c *ledis.Conn) {
		waitBench(c, "LRANGE", "mytestlist", 0, 10)
	}

	bench("lrange10", f)
}

func benchRangeList50() {
	f := func(c *ledis.Conn) {
		waitBench(c, "LRANGE", "mytestlist", 0, 50)
	}

	bench("lrange50", f)
}

func benchRangeList100() {
	f := func(c *ledis.Conn) {
		waitBench(c, "LRANGE", "mytestlist", 0, 100)
	}

	bench("lrange100", f)
}

func benchPopList() {
	f := func(c *ledis.Conn) {
		waitBench(c, "LPOP", "mytestlist")
	}

	bench("lpop", f)
}

var hashSetBase int64 = 0
var hashIncrBase int64 = 0
var hashGetBase int64 = 0
var hashDelBase int64 = 0

func benchHset() {
	f := func(c *ledis.Conn) {
		value := make([]byte, 100)

		n := atomic.AddInt64(&hashSetBase, 1)
		waitBench(c, "HSET", "myhashkey", n, value)
	}

	bench("hset", f)
}

func benchHGet() {
	f := func(c *ledis.Conn) {
		n := atomic.AddInt64(&hashGetBase, 1)
		waitBench(c, "HGET", "myhashkey", n)
	}

	bench("hget", f)
}

func benchHRandGet() {
	f := func(c *ledis.Conn) {
		n := rand.Int() % *number
		waitBench(c, "HGET", "myhashkey", n)
	}

	bench("hrandget", f)
}

func benchHDel() {
	f := func(c *ledis.Conn) {
		n := atomic.AddInt64(&hashDelBase, 1)
		waitBench(c, "HDEL", "myhashkey", n)
	}

	bench("hdel", f)
}

var zsetAddBase int64 = 0
var zsetDelBase int64 = 0
var zsetIncrBase int64 = 0

func benchZAdd() {
	f := func(c *ledis.Conn) {
		member := make([]byte, 16)
		n := atomic.AddInt64(&zsetAddBase, 1)
		waitBench(c, "ZADD", "myzsetkey", n, member)
	}

	bench("zadd", f)
}

func benchZDel() {
	f := func(c *ledis.Conn) {
		n := atomic.AddInt64(&zsetDelBase, 1)
		waitBench(c, "ZREM", "myzsetkey", n)
	}

	bench("zrem", f)
}

func benchZIncr() {
	f := func(c *ledis.Conn) {
		n := atomic.AddInt64(&zsetIncrBase, 1)
		waitBench(c, "ZINCRBY", "myzsetkey", 1, n)
	}

	bench("zincrby", f)
}

func benchZRangeByScore() {
	f := func(c *ledis.Conn) {
		waitBench(c, "ZRANGEBYSCORE", "myzsetkey", 0, rand.Int(), "withscores", "limit", rand.Int()%100, 100)
	}

	bench("zrangebyscore", f)
}

func benchZRangeByRank() {
	f := func(c *ledis.Conn) {
		waitBench(c, "ZRANGE", "myzsetkey", 0, rand.Int()%100)
	}

	bench("zrange", f)
}

func benchZRevRangeByScore() {
	f := func(c *ledis.Conn) {
		waitBench(c, "ZREVRANGEBYSCORE", "myzsetkey", 0, rand.Int(), "withscores", "limit", rand.Int()%100, 100)
	}

	bench("zrevrangebyscore", f)
}

func benchZRevRangeByRank() {
	f := func(c *ledis.Conn) {
		waitBench(c, "ZREVRANGE", "myzsetkey", 0, rand.Int()%100)
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

	for i := 0; i < *clients; i++ {
		c := client.Get()
		c.Close()
	}

	if *round <= 0 {
		*round = 1
	}

	ts := strings.Split(*tests, ",")

	for i := 0; i < *round; i++ {
		for _, s := range ts {
			switch strings.ToLower(s) {
			case "set":
				benchSet()
			case "get":
				benchGet()
			case "randget":
				benchRandGet()
			case "del":
				benchDel()
			case "lpush":
				benchPushList()
			case "lrange":
				benchRangeList10()
				benchRangeList50()
				benchRangeList100()
			case "lpop":
				benchPopList()
			case "hset":
				benchHset()
			case "hget":
				benchHGet()
				benchHRandGet()
			case "hdel":
				benchHDel()
			case "zadd":
				benchZAdd()
			case "zincr":
				benchZIncr()
			case "zrange":
				benchZRangeByRank()
				benchZRangeByScore()
			case "zrevrange":
				//rev is too slow in leveldb, rocksdb or other
				//maybe disable for huge data benchmark
				benchZRevRangeByRank()
				benchZRevRangeByScore()
			case "zdel":
				benchZDel()
			}
		}

		println("")
	}
}
