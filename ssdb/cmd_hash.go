package ssdb

import (
	"github.com/siddontang/golib/hack"
	"strconv"
)

func hsetCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	if n, err := c.app.hash_set(args[0], args[1], args[2]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func hgetCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if v, err := c.app.hash_get(args[0], args[1]); err != nil {
		return err
	} else {
		c.writeBulk(v)
	}

	return nil
}

func hexistsCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	var n int64 = 1
	if v, err := c.app.hash_get(args[0], args[1]); err != nil {
		return err
	} else {
		if v == nil {
			n = 0
		}

		c.writeInteger(n)
	}
	return nil
}

func hdelCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.app.hash_del(args[0], args[1:]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func hlenCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.app.hash_len(args[0]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func hincrbyCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	delta, err := strconv.ParseInt(hack.String(args[2]), 10, 64)
	if err != nil {
		return err
	}

	var n int64
	if n, err = c.app.hash_incrby(args[0], args[1], delta); err != nil {
		return err
	} else {
		c.writeInteger(n)
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

	if err := c.app.hash_mset(args[0], args[1:]); err != nil {
		return err
	} else {
		c.writeStatus(OK)
	}

	return nil
}

func hmgetCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if v, err := c.app.hash_mget(args[0], args[1:]); err != nil {
		return err
	} else {
		c.writeArray(v)
	}

	return nil
}

func hgetallCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.app.hash_getall(args[0]); err != nil {
		return err
	} else {
		c.writeArray(v)
	}

	return nil
}

func hkeysCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.app.hash_keys(args[0]); err != nil {
		return err
	} else {
		c.writeArray(v)
	}

	return nil
}

func hvalsCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.app.hash_values(args[0]); err != nil {
		return err
	} else {
		c.writeArray(v)
	}

	return nil
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
}
