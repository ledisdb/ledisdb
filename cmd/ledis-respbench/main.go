package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/siddontang/go/arena"
	"github.com/siddontang/ledisdb/server"
	"net"
	"runtime"
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

func run(c net.Conn) {
	//buf := make([]byte, 10240)
	ok := []byte("+OK\r\n")
	data := []byte("$4096\r\n")
	data = append(data, make([]byte, 4096)...)
	data = append(data, "\r\n"...)

	var rt time.Duration
	var wt time.Duration

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
			wb.Write(ok)
		case "GET":
			wb.Write(data)
		default:
			wb.WriteString(fmt.Sprintf("-Err %s Not Supported Now", req[0]))
		}

		wb.Flush()

		t3 := time.Now()
		wt += t3.Sub(t2)
	}

	fmt.Printf("rt:%s wt:%s\n", rt.String(), wt.String())
}
