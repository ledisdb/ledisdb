package server

func slaveofCommand(c *client) error {
	if len(c.args) > 1 {
		return ErrCmdParams
	}

	master := ""
	if len(c.args) == 1 {
		master = string(c.args[0])
	}

	if err := c.app.slaveof(master); err != nil {
		return err
	}

	c.writeStatus(OK)

	return nil
}
