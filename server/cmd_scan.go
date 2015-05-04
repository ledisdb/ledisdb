package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/ledis"
)

func parseScanArgs(args [][]byte) (cursor []byte, match string, count int, desc bool, err error) {
	cursor = args[0]

	args = args[1:]

	count = 10

	desc = false

	for i := 0; i < len(args); {
		switch strings.ToUpper(hack.String(args[i])) {
		case "MATCH":
			if i+1 >= len(args) {
				err = ErrCmdParams
				return
			}

			match = hack.String(args[i+1])
			i++
		case "COUNT":
			if i+1 >= len(args) {
				err = ErrCmdParams
				return
			}

			count, err = strconv.Atoi(hack.String(args[i+1]))
			if err != nil {
				return
			}

			i++
		case "ASC":
			desc = false
		case "DESC":
			desc = true
		default:
			err = fmt.Errorf("invalid argument %s", args[i])
			return
		}

		i++
	}

	return
}

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

	cursor, match, count, desc, err := parseScanArgs(args[1:])

	if err != nil {
		return err
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

// XHSCAN key cursor [MATCH match] [COUNT count] [ASC|DESC]
func xhscanCommand(c *client) error {
	args := c.args

	if len(args) < 2 {
		return ErrCmdParams
	}

	key := args[0]

	cursor, match, count, desc, err := parseScanArgs(args[1:])

	if err != nil {
		return err
	}

	var ay []ledis.FVPair

	if !desc {
		ay, err = c.db.HScan(key, cursor, count, false, match)
	} else {
		ay, err = c.db.HRevScan(key, cursor, count, false, match)
	}

	if err != nil {
		return err
	}

	data := make([]interface{}, 2)
	if len(ay) < count {
		data[0] = []byte("")
	} else {
		data[0] = ay[len(ay)-1].Field
	}

	vv := make([][]byte, 0, len(ay)*2)

	for _, v := range ay {
		vv = append(vv, v.Field, v.Value)
	}

	data[1] = vv

	c.resp.writeArray(data)
	return nil
}

// XSSCAN key cursor [MATCH match] [COUNT count] [ASC|DESC]
func xsscanCommand(c *client) error {
	args := c.args

	if len(args) < 2 {
		return ErrCmdParams
	}

	key := args[0]

	cursor, match, count, desc, err := parseScanArgs(args[1:])

	if err != nil {
		return err
	}

	var ay [][]byte

	if !desc {
		ay, err = c.db.SScan(key, cursor, count, false, match)
	} else {
		ay, err = c.db.SRevScan(key, cursor, count, false, match)
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

// XZSCAN key cursor [MATCH match] [COUNT count] [ASC|DESC]
func xzscanCommand(c *client) error {
	args := c.args

	if len(args) < 2 {
		return ErrCmdParams
	}

	key := args[0]

	cursor, match, count, desc, err := parseScanArgs(args[1:])

	if err != nil {
		return err
	}

	var ay []ledis.ScorePair

	if !desc {
		ay, err = c.db.ZScan(key, cursor, count, false, match)
	} else {
		ay, err = c.db.ZRevScan(key, cursor, count, false, match)
	}

	if err != nil {
		return err
	}

	data := make([]interface{}, 2)
	if len(ay) < count {
		data[0] = []byte("")
	} else {
		data[0] = ay[len(ay)-1].Member
	}

	vv := make([][]byte, 0, len(ay)*2)

	for _, v := range ay {
		vv = append(vv, v.Member, num.FormatInt64ToSlice(v.Score))
	}

	data[1] = vv

	c.resp.writeArray(data)
	return nil
}

func init() {
	register("xscan", xscanCommand)
	register("xhscan", xhscanCommand)
	register("xsscan", xsscanCommand)
	register("xzscan", xzscanCommand)
}
