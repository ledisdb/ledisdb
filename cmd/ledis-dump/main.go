package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/server"
	"net"
	"os"
)

var host = flag.String("host", "127.0.0.1", "ledis server host")
var port = flag.Int("port", 6380, "ledis server port")
var sock = flag.String("sock", "", "ledis unix socket domain")
var dumpFile = flag.String("o", "./ledis.dump", "dump file to save")

var fullSyncCmd = []byte("*1\r\n$8\r\nfullsync\r\n") //fullsync

func main() {
	flag.Parse()

	var c net.Conn
	var err error
	var f *os.File

	if f, err = os.OpenFile(*dumpFile, os.O_CREATE|os.O_WRONLY, os.ModePerm); err != nil {
		println(err.Error())
		return
	}

	defer f.Close()

	if len(*sock) != 0 {
		c, err = net.Dial("unix", *sock)
	} else {
		addr := fmt.Sprintf("%s:%d", *host, *port)
		c, err = net.Dial("tcp", addr)
	}

	if err != nil {
		println(err.Error())
		return
	}

	defer c.Close()

	println("dump begin")

	if _, err = c.Write(fullSyncCmd); err != nil {
		println(err.Error())
		return
	}

	rb := bufio.NewReaderSize(c, 16*1024)

	if err = server.ReadBulkTo(rb, f); err != nil {
		println(err.Error())
		return
	}

	println("dump end")
}
