package ssdb

import ()

type CommandFunc func(c *client, args [][]byte) (interface{}, error)

var regCmds = map[string]CommandFunc{}
