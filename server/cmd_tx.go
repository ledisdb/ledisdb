package server

import (
	"errors"
)

var errTxMiss = errors.New("transaction miss")

func beginCommand(c *client) error {
	tx, err := c.db.Begin()
	if err == nil {
		c.tx = tx
		c.db = tx.DB
		c.resp.writeStatus(OK)
	}

	return err
}

func commitCommand(c *client) error {
	if c.tx == nil {
		return errTxMiss
	}

	err := c.tx.Commit()
	c.db, _ = c.ldb.Select(c.tx.Index())
	c.tx = nil

	if err == nil {
		c.resp.writeStatus(OK)
	}

	return err
}

func rollbackCommand(c *client) error {
	if c.tx == nil {
		return errTxMiss
	}

	err := c.tx.Rollback()

	c.db, _ = c.ldb.Select(c.tx.Index())
	c.tx = nil

	if err == nil {
		c.resp.writeStatus(OK)
	}

	return err
}

func init() {
	register("begin", beginCommand)
	register("commit", commitCommand)
	register("rollback", rollbackCommand)
}
