package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/siddontang/goredis"
)

func CmdDump() {
	var host = flag.String("host", "127.0.0.1", "ledis server host")
	var dumpPort = flag.Int("port", 6380, "ledis server port")
	var sock = flag.String("sock", "", "ledis unix socket domain")
	var dumpFile = flag.String("o", "./ledis.dump", "dump file to save")

	flag.Parse()

	var err error
	var f *os.File

	if f, err = os.OpenFile(*dumpFile, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		println(err.Error())
		return
	}

	defer f.Close()

	var addr string
	if len(*sock) != 0 {
		addr = *sock
	} else {
		addr = fmt.Sprintf("%s:%d", *host, *dumpPort)
	}

	c, err := goredis.ConnectWithSize(addr, 16*1024, 4096)
	if err != nil {
		println(err.Error())
		return
	}

	defer c.Close()

	println("dump begin")

	if err = c.Send("fullsync"); err != nil {
		println(err.Error())
		return
	}

	if err = c.ReceiveBulkTo(f); err != nil {
		println(err.Error())
		return
	}

	println("dump end")
}
