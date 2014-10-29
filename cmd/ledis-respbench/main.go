package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/server"
	"net"
	"runtime"
	"sync"
	"time"
)

var addr = flag.String("addr", ":6380", "listen addr")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	l, err := net.Listen("tcp", *addr)

	println("listen", *addr)

	if err != nil {
		fmt.Println(err.Error())
		return
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

var m = map[string][]byte{}
var mu sync.Mutex

func run(c net.Conn) {
	//buf := make([]byte, 10240)
	ok := []byte("+OK\r\n")

	var rt time.Duration
	var wt time.Duration

	rb := bufio.NewReaderSize(c, 10240)
	wb := bufio.NewWriterSize(c, 10240)

	for {
		t1 := time.Now()

		req, err := server.ReadRequest(rb)

		if err != nil {
			break
		}
		t2 := time.Now()

		rt += t2.Sub(t1)

		cmd := string(bytes.ToUpper(req[0]))
		switch cmd {
		case "SET":
			mu.Lock()
			m[string(req[1])] = req[2]
			mu.Unlock()
			wb.Write(ok)
		case "GET":
			mu.Lock()
			v := m[string(req[1])]
			mu.Unlock()
			if v == nil {
				wb.WriteString("$-1\r\n")
			} else {
				wb.WriteString(fmt.Sprintf("$%d\r\n", len(v)))
				wb.Write(v)
				wb.WriteString("\r\n")
			}
		default:
			wb.WriteString(fmt.Sprintf("-Err %s Not Supported Now", req[0]))
		}

		wb.Flush()

		t3 := time.Now()
		wt += t3.Sub(t2)
	}

	fmt.Printf("rt:%s wt:%s\n", rt.String(), wt.String())
}
