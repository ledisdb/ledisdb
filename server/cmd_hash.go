package server

import (
	"github.com/siddontang/ledisdb/ledis"
)

func hsetCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	if n, err := c.db.HSet(args[0], args[1], args[2]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func hgetCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if v, err := c.db.HGet(args[0], args[1]); err != nil {
		return err
	} else {
		c.resp.writeBulk(v)
	}

	return nil
}

func hexistsCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	var n int64 = 1
	if v, err := c.db.HGet(args[0], args[1]); err != nil {
		return err
	} else {
		if v == nil {
			n = 0
		}

		c.resp.writeInteger(n)
	}
	return nil
}

func hdelCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.db.HDel(args[0], args[1:]...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func hlenCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.HLen(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func hincrbyCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	delta, err := ledis.StrInt64(args[2], nil)
	if err != nil {
		return ErrValue
	}

	var n int64
	if n, err = c.db.HIncrBy(args[0], args[1], delta); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func hmsetCommand(c *client) error {
	args := c.args
	if len(args) < 3 {
		return ErrCmdParams
	}

	if len(args[1:])%2 != 0 {
		return ErrCmdParams
	}

	key := args[0]

	args = args[1:]

	kvs := make([]ledis.FVPair, len(args)/2)
	for i := 0; i < len(kvs); i++ {
		kvs[i].Field = args[2*i]
		kvs[i].Value = args[2*i+1]
	}

	if err := c.db.HMset(key, kvs...); err != nil {
		return err
	} else {
		c.resp.writeStatus(OK)
	}

	return nil
}

func hmgetCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if v, err := c.db.HMget(args[0], args[1:]...); err != nil {
		return err
	} else {
		c.resp.writeSliceArray(v)
	}

	return nil
}

func hgetallCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.HGetAll(args[0]); err != nil {
		return err
	} else {
		c.resp.writeFVPairArray(v)
	}

	return nil
}

func hkeysCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.HKeys(args[0]); err != nil {
		return err
	} else {
		c.resp.writeSliceArray(v)
	}

	return nil
}

func hvalsCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.HValues(args[0]); err != nil {
		return err
	} else {
		c.resp.writeSliceArray(v)
	}

	return nil
}

func hclearCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.HClear(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func hmclearCommand(c *client) error {
	args := c.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	if n, err := c.db.HMclear(args...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func hexpireCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.HExpire(args[0], duration); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func hexpireAtCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.HExpireAt(args[0], when); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func httlCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.HTTL(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func hpersistCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.HPersist(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func hxscanCommand(c *client) error {
	return xscanGeneric(c, c.db.HScan)
}

func hxrevscanCommand(c *client) error {
	return xscanGeneric(c, c.db.HRevScan)
}

func init() {
	register("hdel", hdelCommand)
	register("hexists", hexistsCommand)
	register("hget", hgetCommand)
	register("hgetall", hgetallCommand)
	register("hincrby", hincrbyCommand)
	register("hkeys", hkeysCommand)
	register("hlen", hlenCommand)
	register("hmget", hmgetCommand)
	register("hmset", hmsetCommand)
	register("hset", hsetCommand)
	register("hvals", hvalsCommand)

	//ledisdb special command

	register("hclear", hclearCommand)
	register("hmclear", hmclearCommand)
	register("hexpire", hexpireCommand)
	register("hexpireat", hexpireAtCommand)
	register("httl", httlCommand)
	register("hpersist", hpersistCommand)
	register("hxscan", hxscanCommand)
	register("hxrevscan", hxrevscanCommand)
}
