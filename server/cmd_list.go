package server

import (
	"strconv"
	"time"

	"bytes"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/ledisdb/ledis"
)

func lpushCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.db.LPush(args[0], args[1:]...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func rpushCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.db.RPush(args[0], args[1:]...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func lpopCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.LPop(args[0]); err != nil {
		return err
	} else {
		c.resp.writeBulk(v)
	}

	return nil
}

func rpopCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.RPop(args[0]); err != nil {
		return err
	} else {
		c.resp.writeBulk(v)
	}

	return nil
}

func llenCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.LLen(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func lindexCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	index, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.LIndex(args[0], int32(index)); err != nil {
		return err
	} else {
		c.resp.writeBulk(v)
	}

	return nil
}

func lrangeCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	var start int64
	var stop int64
	var err error

	start, err = ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	stop, err = ledis.StrInt64(args[2], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.LRange(args[0], int32(start), int32(stop)); err != nil {
		return err
	} else {
		c.resp.writeSliceArray(v)
	}

	return nil
}

func lclearCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.LClear(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func lmclearCommand(c *client) error {
	args := c.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	if n, err := c.db.LMclear(args...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func lexpireCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.LExpire(args[0], duration); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func lexpireAtCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.LExpireAt(args[0], when); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func lttlCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.LTTL(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func lpersistCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.LPersist(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func blpopCommand(c *client) error {
	keys, timeout, err := lParseBPopArgs(c)
	if err != nil {
		return err
	}

	if ay, err := c.db.BLPop(keys, timeout); err != nil {
		return err
	} else {
		c.resp.writeArray(ay)
	}
	return nil
}

func brpopCommand(c *client) error {
	keys, timeout, err := lParseBPopArgs(c)
	if err != nil {
		return err
	}

	if ay, err := c.db.BRPop(keys, timeout); err != nil {
		return err
	} else {
		c.resp.writeArray(ay)
	}
	return nil

}

func lParseBPopArgs(c *client) (keys [][]byte, timeout time.Duration, err error) {
	args := c.args
	if len(args) < 2 {
		err = ErrCmdParams
		return
	}

	var t float64
	if t, err = strconv.ParseFloat(hack.String(args[len(args)-1]), 64); err != nil {
		return
	}

	timeout = time.Duration(t * float64(time.Second))

	keys = args[0 : len(args)-1]
	return
}

func brpoplpushCommand(c *client) error {
	source, dest, timeout, err := lParseBRPoplpushArgs(c)
	if err != nil {
		return err
	}

	var ttl int64 = -1
	if bytes.Compare(source, dest) == 0 {
		var err error
		ttl, err = c.db.LTTL(source)
		if err != nil {
			return err
		}
	}

	ay, err := c.db.BRPop([][]byte{source}, timeout)
	if err != nil {
		return err
	}

	if ay == nil {
		c.resp.writeBulk(nil)
		return nil
	}

	data, ok := ay[1].([]byte)
	if !ok {
		//not sure if this even possible
		return ErrValue
	}
	if _, err := c.db.LPush(dest, data); err != nil {
		c.db.RPush(source, data) //revert pop
		return err
	}

	//reset ttl
	if ttl != -1 {
		c.db.LExpire(source, ttl)
	}

	c.resp.writeBulk(data)
	return nil

}

func lParseBRPoplpushArgs(c *client) (source []byte, dest []byte, timeout time.Duration, err error) {
	args := c.args
	if len(args) != 3 {
		err = ErrCmdParams
		return
	}

	source = args[0]
	dest = args[1]

	var t float64
	if t, err = strconv.ParseFloat(hack.String(args[len(args)-1]), 64); err != nil {
		return
	}

	timeout = time.Duration(t * float64(time.Second))
	return
}

func rpoplpushCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}
	source, dest := args[0], args[1]

	var ttl int64 = -1
	if bytes.Compare(source, dest) == 0 {
		var err error
		ttl, err = c.db.LTTL(source)
		if err != nil {
			return err
		}
	}

	data, err := c.db.RPop(source)
	if err != nil {
		return err
	}

	if data == nil {
		c.resp.writeBulk(nil)
		return nil
	}

	if _, err := c.db.LPush(dest, data); err != nil {
		c.db.RPush(source, data) //revert pop
		return err
	}

	//reset ttl
	if ttl != -1 {
		c.db.LExpire(source, ttl)
	}

	c.resp.writeBulk(data)
	return nil
}

func lkeyexistsCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}
	if n, err := c.db.LKeyExists(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func lTrimCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	var start int64
	var stop int64
	var err error

	start, err = ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}
	stop, err = ledis.StrInt64(args[2], nil)
	if err != nil {
		return ErrValue
	}

	if err := c.db.LTrim(args[0], start, stop); err != nil {
		return err
	} else {
		c.resp.writeStatus(OK)
	}

	return nil
}

func lTrimFrontCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	var trimSize int32
	var err error

	trimSize, err = ledis.StrInt32(args[1], nil)
	if err != nil || trimSize < 0 {
		return ErrValue
	}

	if n, err := c.db.LTrimFront(args[0], trimSize); err != nil {
		return err
	} else {
		c.resp.writeInteger(int64(n))
	}

	return nil
}

func lTrimBackCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	var trimSize int32
	var err error

	trimSize, err = ledis.StrInt32(args[1], nil)
	if err != nil || trimSize < 0 {
		return ErrValue
	}

	if n, err := c.db.LTrimBack(args[0], trimSize); err != nil {
		return err
	} else {
		c.resp.writeInteger(int64(n))
	}

	return nil
}

func init() {
	register("blpop", blpopCommand)
	register("brpop", brpopCommand)
	register("lindex", lindexCommand)
	register("llen", llenCommand)
	register("lpop", lpopCommand)
	register("lrange", lrangeCommand)
	register("lpush", lpushCommand)
	register("rpop", rpopCommand)
	register("rpush", rpushCommand)
	register("brpoplpush", brpoplpushCommand)
	register("rpoplpush", rpoplpushCommand)

	//ledisdb special command

	register("lclear", lclearCommand)
	register("lmclear", lmclearCommand)
	register("lexpire", lexpireCommand)
	register("lexpireat", lexpireAtCommand)
	register("lttl", lttlCommand)
	register("lpersist", lpersistCommand)
	register("lkeyexists", lkeyexistsCommand)

	register("ltrim_front", lTrimFrontCommand)
	register("ltrim_back", lTrimBackCommand)
	register("ltrim", lTrimCommand)
}
