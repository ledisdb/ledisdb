package server

import (
	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/num"

	"strconv"
	"strings"
	"time"
)

func pingCommand(c *client) error {
	c.resp.writeStatus(PONG)
	return nil
}

func echoCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	c.resp.writeBulk(c.args[0])
	return nil
}

func selectCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	if index, err := strconv.Atoi(hack.String(c.args[0])); err != nil {
		return err
	} else {
		if c.db.IsInMulti() {
			if err := c.script.Select(index); err != nil {
				return err
			} else {
				c.db = c.script.DB
			}
		} else {
			if db, err := c.ldb.Select(index); err != nil {
				return err
			} else {
				c.db = db
			}
		}
		c.resp.writeStatus(OK)
	}

	return nil
}

func infoCommand(c *client) error {
	if len(c.args) > 1 {
		return ErrCmdParams
	}
	var section string
	if len(c.args) == 1 {
		section = strings.ToLower(hack.String(c.args[0]))
	}

	buf := c.app.info.Dump(section)
	c.resp.writeBulk(buf)

	return nil
}

func flushallCommand(c *client) error {
	err := c.ldb.FlushAll()
	if err != nil {
		return err
	}

	//we will restart the replication from master if possible
	c.app.tryReSlaveof()

	c.resp.writeStatus(OK)
	return nil
}

func flushdbCommand(c *client) error {
	_, err := c.db.FlushAll()
	if err != nil {
		return err
	}

	c.resp.writeStatus(OK)
	return nil
}

func timeCommand(c *client) error {
	if len(c.args) != 0 {
		return ErrCmdParams
	}

	t := time.Now()

	//seconds
	s := t.Unix()
	n := t.UnixNano()

	//micro seconds
	m := (n - s*1e9) / 1e3

	ay := []interface{}{
		num.FormatInt64ToSlice(s),
		num.FormatInt64ToSlice(m),
	}

	c.resp.writeArray(ay)
	return nil
}

func configCommand(c *client) error {
	if len(c.args) < 1 {
		return ErrCmdParams
	}

	switch strings.ToLower(hack.String(c.args[0])) {
	case "rewrite":
		if err := c.app.cfg.Rewrite(); err != nil {
			return err
		} else {
			c.resp.writeStatus(OK)
			return nil
		}
	default:
		return ErrCmdParams
	}
}

func init() {
	register("ping", pingCommand)
	register("echo", echoCommand)
	register("select", selectCommand)
	register("info", infoCommand)
	register("flushall", flushallCommand)
	register("flushdb", flushdbCommand)
	register("time", timeCommand)
	register("config", configCommand)
}
