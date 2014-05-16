package server

import (
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
		c.writeInteger(n)
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
		c.writeInteger(n)
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
		c.writeBulk(v)
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
		c.writeBulk(v)
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
		c.writeInteger(n)
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
		return err
	}

	if v, err := c.db.LIndex(args[0], int32(index)); err != nil {
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

	start, err = ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	stop, err = ledis.StrInt64(args[2], nil)
	if err != nil {
		return err
	}

	if v, err := c.db.LRange(args[0], int32(start), int32(stop)); err != nil {
		return err
	} else {
		c.writeArray(v)
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
		c.writeInteger(n)
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

	//ledisdb special command

	register("lclear", lclearCommand)

}
