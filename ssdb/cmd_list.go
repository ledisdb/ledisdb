package ssdb

func lpushCommand(c *client) error {
	return nil
}

func rpushCommand(c *client) error {
	return nil
}

func lpopCommand(c *client) error {
	return nil
}

func rpopCommand(c *client) error {
	return nil
}

func llenCommand(c *client) error {
	return nil
}

func lindexCommand(c *client) error {
	return nil
}

func lrangeCommand(c *client) error {
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
