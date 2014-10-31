package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/siddontang/go/arena"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/server"
	"net"
	"os"
	"runtime"
	"time"
)

var KB = config.KB
var MB = config.MB
var GB = config.GB

var addr = flag.String("addr", ":6380", "listen addr")
var name = flag.String("db_name", "", "db name")

var ldb *ledis.Ledis
var db *ledis.DB

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
	l, err := net.Listen("tcp", *addr)

	println("listen", *addr)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if len(*name) > 0 {
		cfg := config.NewConfigDefault()
		cfg.DataDir = "./var/ledis_respbench"
		cfg.DBName = *name
		os.RemoveAll(cfg.DBPath)
		defer os.RemoveAll(cfg.DBPath)

		os.MkdirAll(cfg.DBPath, 0755)

		cfg.LevelDB.BlockSize = 32 * KB
		cfg.LevelDB.CacheSize = 512 * MB
		cfg.LevelDB.WriteBufferSize = 64 * MB
		cfg.LevelDB.MaxOpenFiles = 1000

		setRocksDB(&cfg.RocksDB)

		ldb, err = ledis.Open(cfg)
		if err != nil {
			println(err.Error())
			return
		}

		db, _ = ldb.Select(0)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			println(err.Error())
			continue
		}
		go run(c)
	}
}

func run(c net.Conn) {
	//buf := make([]byte, 10240)
	ok := []byte("+OK\r\n")
	data := []byte("$4096\r\n")
	data = append(data, make([]byte, 4096)...)
	data = append(data, "\r\n"...)

	var rt time.Duration
	var wt time.Duration
	var st time.Duration
	var gt time.Duration

	rb := bufio.NewReaderSize(c, 10240)
	wb := bufio.NewWriterSize(c, 10240)

	a := arena.NewArena(10240)

	for {
		t1 := time.Now()

		a.Reset()

		req, err := server.ReadRequest(rb, a)

		if err != nil {
			break
		}
		t2 := time.Now()

		rt += t2.Sub(t1)

		cmd := string(bytes.ToUpper(req[0]))
		switch cmd {
		case "SET":
			if db != nil {
				db.Set(req[1], req[2])
				st += time.Now().Sub(t2)
			}
			wb.Write(ok)
		case "GET":
			if db != nil {
				d, _ := db.GetSlice(req[1])
				gt += time.Now().Sub(t2)
				if d == nil {
					wb.Write(data)
				} else {
					wb.WriteString(fmt.Sprintf("$%d\r\n", d.Size()))
					wb.Write(d.Data())
					wb.WriteString("\r\n")
					d.Free()
				}
			} else {
				wb.Write(data)
			}
		default:
			wb.WriteString(fmt.Sprintf("-Err %s Not Supported Now\r\n", req[0]))
		}

		wb.Flush()

		t3 := time.Now()
		wt += t3.Sub(t2)
	}

	fmt.Printf("rt:%s wt %s, gt:%s, st:%s\n", rt.String(), wt.String(), gt.String(), st.String())
}
