package ledis

import ()

func hsetCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	if n, err := c.db.HSet(args[0], args[1], args[2]); err != nil {
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

	if v, err := c.db.HGet(args[0], args[1]); err != nil {
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
	if v, err := c.db.HGet(args[0], args[1]); err != nil {
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

	if n, err := c.db.HDel(args[0], args[1:]); err != nil {
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

	if n, err := c.db.HLen(args[0]); err != nil {
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

	delta, err := StrInt64(args[2], nil)
	if err != nil {
		return err
	}

	var n int64
	if n, err = c.db.HIncrBy(args[0], args[1], delta); err != nil {
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

	if err := c.db.HMset(args[0], args[1:]); err != nil {
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

	if v, err := c.db.HMget(args[0], args[1:]); err != nil {
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

	if v, err := c.db.HGetAll(args[0]); err != nil {
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

	if v, err := c.db.HKeys(args[0]); err != nil {
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

	if v, err := c.db.HValues(args[0]); err != nil {
		return err
	} else {
		c.writeArray(v)
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
		c.writeInteger(n)
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

	//ledisdb special command

	register("hclear", hclearCommand)
}
