package ssdb

import (
	"github.com/siddontang/golib/hack"
	"strconv"
)

func lpushCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.app.list_lpush(args[0], args[1:]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func rpushCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.app.list_rpush(args[0], args[1:]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func lpopCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.app.list_lpop(args[0]); err != nil {
		return err
	} else {
		c.writeBulk(v)
	}

	return nil
}

func rpopCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.app.list_rpop(args[0]); err != nil {
		return err
	} else {
		c.writeBulk(v)
	}

	return nil
}

func llenCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.app.list_len(args[0]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func lindexCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	index, err := strconv.ParseInt(hack.String(args[1]), 10, 64)
	if err != nil {
		return err
	}

	if v, err := c.app.list_index(args[0], int32(index)); err != nil {
		return err
	} else {
		c.writeBulk(v)
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

	start, err = strconv.ParseInt(hack.String(args[1]), 10, 64)
	if err != nil {
		return err
	}

	stop, err = strconv.ParseInt(hack.String(args[2]), 10, 64)
	if err != nil {
		return err
	}

	if v, err := c.app.list_range(args[0], int32(start), int32(stop)); err != nil {
		return err
	} else {
		c.writeArray(v)
	}

	return nil
}

func init() {
	register("lindex", lindexCommand)
	register("llen", llenCommand)
	register("lpop", lpopCommand)
	register("lrange", lrangeCommand)
	register("lpush", lpushCommand)
	register("rpop", rpopCommand)
	register("rpush", rpushCommand)
}
