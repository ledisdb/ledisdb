package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"

	"strings"
)

type CommandFunc func(c *client) error

var regCmds = map[string]CommandFunc{}

func register(name string, f CommandFunc) {
	if _, ok := regCmds[strings.ToLower(name)]; ok {
		panic(fmt.Sprintf("%s has been registered", name))
	}

	regCmds[name] = f
}

func pingCommand(c *client) error {
	c.writeStatus(PONG)
	return nil
}

func echoCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	c.writeBulk(c.args[0])
	return nil
}

func selectCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	if index, err := strconv.Atoi(ledis.String(c.args[0])); err != nil {
		return err
	} else {
		if db, err := c.ldb.Select(index); err != nil {
			return err
		} else {
			c.db = db
			c.writeStatus(OK)
		}
	}
	return nil
}

func init() {
	register("ping", pingCommand)
	register("echo", echoCommand)
	register("select", selectCommand)
}
