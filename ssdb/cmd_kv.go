package ssdb

func getCommand(c *client) error {
	return nil
}

func setCommand(c *client) error {
	return nil
}

func getsetCommand(c *client) error {
	return nil
}

func setnxCommand(c *client) error {
	return nil
}

func existsCommand(c *client) error {
	return nil
}

func incrCommand(c *client) error {
	return nil
}

func decrCommand(c *client) error {
	return nil
}

func incrbyCommand(c *client) error {
	return nil
}

func decrbyCommand(c *client) error {
	return nil
}

func delCommand(c *client) error {
	return nil
}

func msetCommand(c *client) error {
	return nil
}

func setexCommand(c *client) error {
	return nil
}

func mgetCommand(c *client) error {
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
	register("setex", setexCommand)
	register("setnx", setnxCommand)
}
