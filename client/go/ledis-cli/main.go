package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"os"
	"strings"
)

var ip = flag.String("h", "127.0.0.1", "ledisdb server ip (default 127.0.0.1)")
var port = flag.Int("p", 6380, "ledisdb server port (default 6380)")
var socket = flag.String("s", "", "ledisdb server socket, overwrite ip and port")

func main() {
	flag.Parse()

	cfg := new(ledis.Config)
	if len(*socket) > 0 {
		cfg.Addr = *socket
	} else {
		cfg.Addr = fmt.Sprintf("%s:%d", *ip, *port)
	}

	cfg.MaxIdleConns = 1

	c := ledis.NewClient(cfg)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("ledis %s > ", cfg.Addr)

		cmd, _ := reader.ReadString('\n')

		cmds := strings.Fields(cmd)
		if len(cmds) == 0 {
			continue
		} else {
			args := make([]interface{}, len(cmds[1:]))
			for i := range args {
				args[i] = cmds[1+i]
			}
			r, err := c.Do(cmds[0], args...)
			if err != nil {
				fmt.Printf("%s", err.Error())
			} else {
				printReply(r)
			}

			fmt.Printf("\n")
		}
	}
}

func printReply(reply interface{}) {
	switch reply := reply.(type) {
	case int64:
		fmt.Printf("(integer) %d", reply)
	case string:
		fmt.Printf("%q", reply)
	case []byte:
		fmt.Printf("%q", reply)
	case nil:
		fmt.Printf("(empty list or set)")
	case ledis.Error:
		fmt.Printf("%s", string(reply))
	case []interface{}:
		for i, v := range reply {
			fmt.Printf("%d) ", i)
			if v == nil {
				fmt.Printf("(nil)")
			} else {
				fmt.Printf("%q", v)
			}
		}
	default:
		fmt.Printf("invalid ledis reply")
	}
}
