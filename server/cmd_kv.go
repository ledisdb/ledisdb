package server

import (
	"github.com/siddontang/ledisdb/ledis"
)

func getCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.Get(args[0]); err != nil {
		return err
	} else {
		c.resp.writeBulk(v)
	}
	return nil
}

func setCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if err := c.db.Set(args[0], args[1]); err != nil {
		return err
	} else {
		c.resp.writeStatus(OK)
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
		c.resp.writeBulk(v)
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
		c.resp.writeInteger(n)
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
		c.resp.writeInteger(n)
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
		c.resp.writeInteger(n)
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
		c.resp.writeInteger(n)
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
		c.resp.writeInteger(n)
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
		c.resp.writeInteger(n)
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
		c.resp.writeInteger(n)
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
		c.resp.writeStatus(OK)
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
		c.resp.writeSliceArray(v)
	}

	return nil
}

func expireCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := c.db.Expire(args[0], duration); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func expireAtCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := c.db.ExpireAt(args[0], when); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func ttlCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.TTL(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func persistCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.Persist(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

// func (db *DB) Expire(key []byte, duration int6
// func (db *DB) ExpireAt(key []byte, when int64)
// func (db *DB) TTL(key []byte) (int64, error)

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
	register("expire", expireCommand)
	register("expireat", expireAtCommand)
	register("ttl", ttlCommand)
	register("persist", persistCommand)
}
