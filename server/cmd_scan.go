package server

import (
	"fmt"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"
	"strings"
)

// XSCAN type cursor [MATCH match] [COUNT count] [ASC|DESC]
func xscanCommand(c *client) error {
	args := c.args

	if len(args) < 2 {
		return ErrCmdParams
	}

	var dataType ledis.DataType
	switch strings.ToUpper(hack.String(args[0])) {
	case "KV":
		dataType = ledis.KV
	case "HASH":
		dataType = ledis.HASH
	case "LIST":
		dataType = ledis.LIST
	case "SET":
		dataType = ledis.SET
	case "ZSET":
		dataType = ledis.ZSET
	default:
		return fmt.Errorf("invalid key type %s", args[0])
	}

	cursor := args[1]

	args = args[2:]

	match := ""
	count := 10

	desc := false

	var err error

	for i := 0; i < len(args); {
		switch strings.ToUpper(hack.String(args[i])) {
		case "MATCH":
			if i+1 >= len(args) {
				return ErrCmdParams
			}

			match = hack.String(args[i+1])
			i = i + 2
		case "COUNT":
			if i+1 >= len(args) {
				return ErrCmdParams
			}

			count, err = strconv.Atoi(hack.String(args[i+1]))
			if err != nil {
				return err
			}

			i = i + 2
		case "ASC":
			desc = false
			i++
		case "DESC":
			desc = true
			i++
		default:
			return fmt.Errorf("invalid argument %s", args[i])
		}
	}

	var ay [][]byte
	if !desc {
		ay, err = c.db.Scan(dataType, cursor, count, false, match)
	} else {
		ay, err = c.db.RevScan(dataType, cursor, count, false, match)
	}

	if err != nil {
		return err
	}

	data := make([]interface{}, 2)
	if len(ay) < count {
		data[0] = []byte("")
	} else {
		data[0] = ay[len(ay)-1]
	}
	data[1] = ay
	c.resp.writeArray(data)
	return nil
}

func init() {
	register("xscan", xscanCommand)
}
