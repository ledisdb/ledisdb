package server

import (
	"fmt"
	goledis "github.com/siddontang/ledisdb/client/go/ledis"
	"github.com/siddontang/ledisdb/ledis"
	"strings"
	"time"
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

// unlike redis, restore will try to delete old key first
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

func xdump(db *ledis.DB, tp string, key []byte) ([]byte, error) {
	var err error
	var data []byte
	switch tp {
	case "kv":
		data, err = db.Dump(key)
	case "hash":
		data, err = db.HDump(key)
	case "list":
		data, err = db.LDump(key)
	case "set":
		data, err = db.SDump(key)
	case "zset":
		data, err = db.ZDump(key)
	default:
		err = fmt.Errorf("invalid key type %s", tp)
	}
	return data, err
}

func xdel(db *ledis.DB, tp string, key []byte) error {
	var err error
	switch tp {
	case "kv":
		_, err = db.Del(key)
	case "hash":
		_, err = db.HClear(key)
	case "list":
		_, err = db.LClear(key)
	case "set":
		_, err = db.SClear(key)
	case "zset":
		_, err = db.ZClear(key)
	default:
		err = fmt.Errorf("invalid key type %s", tp)
	}
	return err
}

func xdumpCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	tp := string(args[0])
	key := args[1]

	if data, err := xdump(c.db, tp, key); err != nil {
		return err
	} else {
		c.resp.writeBulk(data)
	}
	return nil
}

//XMIGRATE host port type key destination-db timeout [COPY]
func xmigrateCommand(c *client) error {
	args := c.args

	if len(args) != 6 && len(args) != 7 {
		return ErrCmdParams
	}

	addr := fmt.Sprintf("%s:%d", string(args[0]), string(args[1]))
	tp := string(args[2])
	key := args[3]
	db, err := ledis.StrUint64(args[4], nil)
	if err != nil {
		return err
	} else if db >= uint64(ledis.MaxDBNumber) {
		return fmt.Errorf("invalid db index %d, must < %d", db, ledis.MaxDBNumber)
	}
	var timeout int64
	timeout, err = ledis.StrInt64(args[5], nil)
	if err != nil {
		return err
	} else if timeout < 0 {
		return fmt.Errorf("invalid timeout %d", timeout)
	}

	onlyCopy := false
	if len(args) == 7 {
		if strings.ToUpper(string(args[6])) == "COPY" {
			onlyCopy = true
		}
	}

	var m *ledis.Multi
	if m, err = c.db.Multi(); err != nil {
		return err
	}
	defer m.Close()

	var data []byte
	data, err = xdump(m.DB, tp, key)
	if err != nil {
		return err
	} else if data == nil {
		c.resp.writeStatus(NOKEY)
		return nil
	}

	c.app.migrateConnM.Lock()
	defer c.app.migrateConnM.Unlock()

	conn, ok := c.app.migrateConns[addr]
	if !ok {
		conn = goledis.NewConn(addr)
		c.app.migrateConns[addr] = conn
	}

	//timeout is milliseconds
	conn.SetConnectTimeout(time.Duration(timeout) * time.Millisecond)

	if _, err = conn.Do("restore", key, data); err != nil {
		return err
	}

	if !onlyCopy {
		if err = xdel(m.DB, tp, key); err != nil {
			return err
		}
	}

	c.resp.writeStatus(OK)
	return nil
}

func init() {
	register("dump", dumpCommand)
	register("ldump", ldumpCommand)
	register("hdump", hdumpCommand)
	register("sdump", sdumpCommand)
	register("zdump", zdumpCommand)
	register("restore", restoreCommand)
	register("xdump", xdumpCommand)
	register("xmigrate", xmigrateCommand)
}
