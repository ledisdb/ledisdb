package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"
	"strings"
)

func slaveofCommand(c *client) error {
	args := c.args

	if len(args) != 2 {
		return ErrCmdParams
	}

	masterAddr := ""

	if strings.ToLower(ledis.String(args[0])) == "no" &&
		strings.ToLower(ledis.String(args[1])) == "one" {
		//stop replication, use master = ""
	} else {
		if _, err := strconv.ParseInt(ledis.String(args[1]), 10, 16); err != nil {
			return err
		}

		masterAddr = fmt.Sprintf("%s:%s", args[0], args[1])
	}

	if err := c.app.slaveof(masterAddr); err != nil {
		return err
	}

	c.writeStatus(OK)

	return nil
}
