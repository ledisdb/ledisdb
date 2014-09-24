package server

import (
	"fmt"
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
