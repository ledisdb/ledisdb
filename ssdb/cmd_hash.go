package ssdb

func hsetCommand(c *client) error {
	return nil
}

func hgetCommand(c *client) error {
	return nil
}

func hexistsCommand(c *client) error {
	return nil
}

func hmsetCommand(c *client) error {
	return nil
}

func hdelCommand(c *client) error {
	return nil
}

func hlenCommand(c *client) error {
	return nil
}

func hincrbyCommand(c *client) error {
	return nil
}

func hmgetCommand(c *client) error {
	return nil
}

func hgetallCommand(c *client) error {
	return nil
}

func hkeysCommand(c *client) error {
	return nil
}

func hvalsCommand(c *client) error {
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
