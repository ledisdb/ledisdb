package main

import (
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/client/go/redis"
	"math/rand"
	"sync"
	"time"
)

var ip = flag.String("ip", "127.0.0.1", "redis/ledis/ssdb server ip")
var port = flag.Int("port", 6379, "redis/ledis/ssdb server port")
var number = flag.Int("n", 1000, "request number")
var clients = flag.Int("c", 50, "number of clients")

var wg sync.WaitGroup

var pool *redis.Pool

var loop int = 0

func waitBench(cmd string, args ...interface{}) {
	defer wg.Done()

	c := pool.Get()
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

func benchSet() {
	n := rand.Int()
	f := func() {
		waitBench("set", n, n)
	}

	bench("set", f)
}

func benchGet() {
	n := rand.Int()
	f := func() {
		waitBench("get", n)
	}

	bench("get", f)
}

func benchIncr() {
	n := rand.Int()
	f := func() {
		waitBench("incr", n)
	}

	bench("incr", f)
}

func benchPushList() {
	n := rand.Int()
	f := func() {
		waitBench("rpush", "mytestlist", n)
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

func benchHset() {
	n := rand.Int()

	f := func() {
		waitBench("hset", "myhashkey", n, n)
	}

	bench("hset", f)
}

func benchHIncr() {
	n := rand.Int()
	f := func() {
		waitBench("hincrby", "myhashkey", n, 1)
	}

	bench("hincrby", f)
}

func benchHGet() {
	n := rand.Int()
	f := func() {
		waitBench("hget", "myhashkey", n)
	}

	bench("hget", f)
}

func benchHDel() {
	n := rand.Int()
	f := func() {
		waitBench("hdel", "myhashkey", n)
	}

	bench("hdel", f)
}

func benchZAdd() {
	n := rand.Int()
	f := func() {
		waitBench("zadd", "myzsetkey", n, n)
	}

	bench("zadd", f)
}

func benchZDel() {
	n := rand.Int()
	f := func() {
		waitBench("zrem", "myzsetkey", n)
	}

	bench("zrem", f)
}

func benchZIncr() {
	n := rand.Int()

	f := func() {
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

	f := func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}

		return c, nil
	}
	pool = redis.NewPool(f, *clients)

	benchSet()
	benchIncr()
	benchGet()

	benchPushList()
	benchRangeList10()
	benchRangeList50()
	benchRangeList100()
	benchPopList()

	benchHset()
	benchHGet()
	benchHIncr()
	benchHDel()

	benchZAdd()
	benchZIncr()
	benchZRangeByRank()
	benchZRangeByScore()
	benchZRevRangeByRank()
	benchZRevRangeByScore()
	benchZDel()
}
