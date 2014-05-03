package ssdb

func init() {
	register("zadd", zaddCommand)
	register("zcard", zcardCommand)
	register("zcount", zcountCommand)
	register("zincrby", zincrbyCommand)
	register("zrange", zrangeCommand)
	register("zrangebyscore", zrangebyscoreCommand)
	register("zrank", zrankCommand)
	register("zrem", zremCommand)
	register("zremrangebyrank", zremrangebyrankCommand)
	register("zremrangebyscore", zremrangebyscoreCommand)
	register("zrevrange", zrevrangeCommand)
	register("zrevrank", zrevrankCommand)
	register("zrevrangebyscore", zrevrangebyscoreCommand)
	register("zscore", zscoreCommand)
}

func zcardCommand(c *client) error {
	return nil
}

func zscoreCommand(c *client) error {
	return nil
}

func zremCommand(c *client) error {
	return nil
}

func zrankCommand(c *client) error            { return nil }
func zrevrankCommand(c *client) error         { return nil }
func zcountCommand(c *client) error           { return nil }
func zremrangebyrankCommand(c *client) error  { return nil }
func zremrangebyscoreCommand(c *client) error { return nil }
func zrangeCommand(c *client) error           { return nil }
func zrevrangeCommand(c *client) error        { return nil }
func zaddCommand(c *client) error             { return nil }
func zincrbyCommand(c *client) error          { return nil }
func zrangebyscoreCommand(c *client) error    { return nil }
func zrevrangebyscoreCommand(c *client) error { return nil }
