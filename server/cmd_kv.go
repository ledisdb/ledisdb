package server

import (
	"github.com/siddontang/ledisdb/ledis"
	"strings"
)

func getCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.Get(args[0]); err != nil {
		return err
	} else {
		c.writeBulk(v)
	}
	return nil
}

func setCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if err := c.db.Set(args[0], args[1]); err != nil {
		return err
	} else {
		c.writeStatus(OK)
	}

	return nil
}

func getsetCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if v, err := c.db.GetSet(args[0], args[1]); err != nil {
		return err
	} else {
		c.writeBulk(v)
	}

	return nil
}

func setnxCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if n, err := c.db.SetNX(args[0], args[1]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func existsCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.Exists(args[0]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func incrCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.Incr(c.args[0]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func decrCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.Decr(c.args[0]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func incrbyCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	delta, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if n, err := c.db.IncryBy(c.args[0], delta); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func decrbyCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	delta, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if n, err := c.db.DecrBy(c.args[0], delta); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func delCommand(c *client) error {
	args := c.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	if n, err := c.db.Del(args...); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func msetCommand(c *client) error {
	args := c.args
	if len(args) == 0 || len(args)%2 != 0 {
		return ErrCmdParams
	}

	kvs := make([]ledis.KVPair, len(args)/2)
	for i := 0; i < len(kvs); i++ {
		kvs[i].Key = args[2*i]
		kvs[i].Value = args[2*i+1]
	}

	if err := c.db.MSet(kvs...); err != nil {
		return err
	} else {
		c.writeStatus(OK)
	}

	return nil
}

// func setexCommand(c *client) error {
// 	return nil
// }

func mgetCommand(c *client) error {
	args := c.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	if v, err := c.db.MGet(args...); err != nil {
		return err
	} else {
		c.writeArray(v)
	}

	return nil
}

func scanCommand(c *client) error {
	args := c.args

	if len(args) != 1 && len(args) != 3 {
		return ErrCmdParams
	}
	var offset int64
	var count int64
	var err error

	if offset, err = ledis.StrInt64(args[0], nil); err != nil {
		return err
	}

	//now we only support count
	if len(args) == 3 {
		if strings.ToLower(ledis.String(args[1])) != "count" {
			return ErrCmdParams
		}

		if count, err = ledis.StrInt64(args[2], nil); err != nil {
			return err
		}
	}

	if v, err := c.db.Scan(int(offset), int(count)); err != nil {
		return err
	} else {
		c.writeArray(v)
	}

	return nil
}

func init() {
	register("decr", decrCommand)
	register("decrby", decrbyCommand)
	register("del", delCommand)
	register("exists", existsCommand)
	register("get", getCommand)
	register("getset", getsetCommand)
	register("incr", incrCommand)
	register("incrby", incrbyCommand)
	register("mget", mgetCommand)
	register("mset", msetCommand)
	register("set", setCommand)
	register("setnx", setnxCommand)
	register("scan", scanCommand)
}
