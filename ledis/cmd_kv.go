package ledis

import ()

func getCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.app.kv_get(args[0]); err != nil {
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

	if err := c.app.kv_set(args[0], args[1]); err != nil {
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

	if v, err := c.app.kv_getset(args[0], args[1]); err != nil {
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

	if n, err := c.app.kv_setnx(args[0], args[1]); err != nil {
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

	if n, err := c.app.kv_exists(args[0]); err != nil {
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

	if n, err := c.app.kv_incr(c.args[0], 1); err != nil {
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

	if n, err := c.app.kv_incr(c.args[0], -1); err != nil {
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

	delta, err := StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if n, err := c.app.kv_incr(c.args[0], delta); err != nil {
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

	delta, err := StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if n, err := c.app.kv_incr(c.args[0], -delta); err != nil {
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

	if n, err := c.app.tx_del(args); err != nil {
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

	if err := c.app.tx_mset(args); err != nil {
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

	if v, err := c.app.kv_mget(args); err != nil {
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
	//	register("setex", setexCommand)
	register("setnx", setnxCommand)
}
