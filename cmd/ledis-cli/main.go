package main

import (
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"regexp"
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

	setHistoryCapacity(100)

	reg, _ := regexp.Compile(`'.*?'|".*?"|\S+`)

	for {
		cmd, err := line(fmt.Sprintf("%s> ", cfg.Addr))
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			return
		}

		cmds := reg.FindAllString(cmd, -1)
		if len(cmds) == 0 {
			continue
		} else {
			addHistory(cmd)

			args := make([]interface{}, len(cmds[1:]))
			for i := range args {
				args[i] = strings.Trim(string(cmds[1+i]), "\"'")
			}
			r, err := c.Do(cmds[0], args...)
			if err != nil {
				fmt.Printf("%s", err.Error())
			} else {
				printReply(cmd, r)
			}

			fmt.Printf("\n")
		}
	}
}

func printReply(cmd string, reply interface{}) {
	switch reply := reply.(type) {
	case int64:
		fmt.Printf("(integer) %d", reply)
	case string:
		fmt.Printf("%s", reply)
	case []byte:
		fmt.Printf("%s", string(reply))
	case nil:
		fmt.Printf("(nil)")
	case ledis.Error:
		fmt.Printf("%s", string(reply))
	case []interface{}:
		for i, v := range reply {
			fmt.Printf("%d) ", i+1)
			if v == nil {
				fmt.Printf("(nil)")
			} else {
				fmt.Printf("%q", v)
			}
			if i != len(reply)-1 {
				fmt.Printf("\n")
			}
		}
	default:
		fmt.Printf("invalid ledis reply")
	}
}
