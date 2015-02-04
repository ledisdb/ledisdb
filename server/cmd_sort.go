package server

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func xsort(c *client, tp string, key []byte, offset int, size int, alpha bool,
	desc bool, sortBy []byte, sortGet [][]byte) ([][]byte, error) {
	var ay [][]byte
	var err error
	switch strings.ToUpper(tp) {
	case "LIST":
		ay, err = c.db.XLSort(key, offset, size, alpha, desc, sortBy, sortGet)
	case "SET":
		ay, err = c.db.XSSort(key, offset, size, alpha, desc, sortBy, sortGet)
	case "ZSET":
		ay, err = c.db.XZSort(key, offset, size, alpha, desc, sortBy, sortGet)
	default:
		err = fmt.Errorf("invalid key type %s", tp)
	}
	return ay, err
}

func xlsortCommand(c *client) error {
	return handleXSort(c, "LIST")
}
func xssortCommand(c *client) error {
	return handleXSort(c, "SET")
}
func xzsortCommand(c *client) error {
	return handleXSort(c, "ZSET")
}

var ascArg = []byte("asc")
var descArg = []byte("desc")
var alphaArg = []byte("alpha")
var limitArg = []byte("limit")
var storeArg = []byte("store")
var byArg = []byte("by")
var getArg = []byte("get")

func handleXSort(c *client, tp string) error {
	args := c.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	key := args[0]
	desc := false
	alpha := false
	offset := 0
	size := 0
	var storeKey []byte
	var sortBy []byte
	var sortGet [][]byte
	var err error

	for i := 1; i < len(args); {
		if bytes.EqualFold(args[i], ascArg) {
			desc = false
		} else if bytes.EqualFold(args[i], descArg) {
			desc = true
		} else if bytes.EqualFold(args[i], alphaArg) {
			alpha = true
		} else if bytes.EqualFold(args[i], limitArg) && i+2 < len(args) {
			if offset, err = strconv.Atoi(string(args[i+1])); err != nil {
				return err
			}
			if size, err = strconv.Atoi(string(args[i+2])); err != nil {
				return err
			}
			i = i + 2
		} else if bytes.EqualFold(args[i], storeArg) && i+1 < len(args) {
			storeKey = args[i+1]
			i++
		} else if bytes.EqualFold(args[i], byArg) && i+1 < len(args) {
			sortBy = args[i+1]
			i++
		} else if bytes.EqualFold(args[i], getArg) && i+1 < len(args) {
			sortGet = append(sortGet, args[i+1])
			i++
		} else {
			return ErrCmdParams
		}

		i++
	}

	ay, err := xsort(c, tp, key, offset, size, alpha, desc, sortBy, sortGet)
	if err != nil {
		return err
	}

	if storeKey == nil {
		c.resp.writeSliceArray(ay)
	} else {
		// not threadsafe now, need lock???
		if _, err = c.db.LClear(storeKey); err != nil {
			return err
		}

		if n, err := c.db.RPush(storeKey, ay...); err != nil {
			return err
		} else {
			c.resp.writeInteger(n)
		}
	}
	return nil
}

func init() {
	register("xlsort", xlsortCommand)
	register("xssort", xssortCommand)
	register("xzsort", xzsortCommand)
}
