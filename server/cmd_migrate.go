package server

import (
	"github.com/siddontang/ledisdb/ledis"
)

func dumpCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	key := c.args[0]
	if data, err := c.db.Dump(key); err != nil {
		return err
	} else {
		c.resp.writeBulk(data)
	}

	return nil
}

func ldumpCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	key := c.args[0]
	if data, err := c.db.LDump(key); err != nil {
		return err
	} else {
		c.resp.writeBulk(data)
	}

	return nil
}

func hdumpCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	key := c.args[0]
	if data, err := c.db.HDump(key); err != nil {
		return err
	} else {
		c.resp.writeBulk(data)
	}

	return nil
}

func sdumpCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	key := c.args[0]
	if data, err := c.db.SDump(key); err != nil {
		return err
	} else {
		c.resp.writeBulk(data)
	}

	return nil
}

func zdumpCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	key := c.args[0]
	if data, err := c.db.ZDump(key); err != nil {
		return err
	} else {
		c.resp.writeBulk(data)
	}

	return nil
}

func restoreCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	key := args[0]
	ttl, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}
	data := args[2]

	if err = c.db.Restore(key, ttl, data); err != nil {
		return err
	} else {
		c.resp.writeStatus(OK)
	}

	return nil
}

func init() {
	register("dump", dumpCommand)
	register("ldump", ldumpCommand)
	register("hdump", hdumpCommand)
	register("sdump", sdumpCommand)
	register("zdump", zdumpCommand)
	register("restore", restoreCommand)
}
