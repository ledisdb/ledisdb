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
	switch strings.ToUpper(tp) {
	case "KV":
		data, err = db.Dump(key)
	case "HASH":
		data, err = db.HDump(key)
	case "LIST":
		data, err = db.LDump(key)
	case "SET":
		data, err = db.SDump(key)
	case "ZSET":
		data, err = db.ZDump(key)
	default:
		err = fmt.Errorf("invalid key type %s", tp)
	}
	return data, err
}

func xdel(db *ledis.DB, tp string, key []byte) error {
	var err error
	switch strings.ToUpper(tp) {
	case "KV":
		_, err = db.Del(key)
	case "HASH":
		_, err = db.HClear(key)
	case "LIST":
		_, err = db.LClear(key)
	case "SET":
		_, err = db.SClear(key)
	case "ZSET":
		_, err = db.ZClear(key)
	default:
		err = fmt.Errorf("invalid key type %s", tp)
	}
	return err
}

func xttl(db *ledis.DB, tp string, key []byte) (int64, error) {
	switch strings.ToUpper(tp) {
	case "KV":
		return db.TTL(key)
	case "HASH":
		return db.HTTL(key)
	case "LIST":
		return db.LTTL(key)
	case "SET":
		return db.STTL(key)
	case "ZSET":
		return db.ZTTL(key)
	default:
		return 0, fmt.Errorf("invalid key type %s", tp)
	}
}

func xscan(db *ledis.DB, tp string, count int) ([][]byte, error) {
	switch strings.ToUpper(tp) {
	case "KV":
		return db.Scan(nil, count, false, "")
	case "HASH":
		return db.HScan(nil, count, false, "")
	case "LIST":
		return db.LScan(nil, count, false, "")
	case "SET":
		return db.SScan(nil, count, false, "")
	case "ZSET":
		return db.ZScan(nil, count, false, "")
	default:
		return nil, fmt.Errorf("invalid key type %s", tp)
	}
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

func (app *App) getMigrateClient(addr string) *goledis.Client {
	app.migrateM.Lock()

	mc, ok := app.migrateClients[addr]
	if !ok {
		mc = goledis.NewClient(&goledis.Config{addr, 4, 0, 0})
		app.migrateClients[addr] = mc

	}

	app.migrateM.Unlock()

	return mc
}

//XMIGRATEDB host port tp count db timeout
//select count tp type keys and migrate
//will block any other write operations
//maybe only for xcodis
func xmigratedbCommand(c *client) error {
	args := c.args
	if len(args) != 6 {
		return ErrCmdParams
	}

	addr := fmt.Sprintf("%s:%s", string(args[0]), string(args[1]))
	if addr == c.app.cfg.Addr {
		//same server， can not migrate
		return fmt.Errorf("migrate in same server is not allowed")
	}

	tp := string(args[2])

	count, err := ledis.StrInt64(args[3], nil)
	if err != nil {
		return err
	} else if count <= 0 {
		count = 10
	}

	db, err := ledis.StrUint64(args[4], nil)
	if err != nil {
		return err
	} else if db >= uint64(ledis.MaxDBNumber) {
		return fmt.Errorf("invalid db index %d, must < %d", db, ledis.MaxDBNumber)
	}

	timeout, err := ledis.StrInt64(args[5], nil)
	if err != nil {
		return err
	} else if timeout < 0 {
		return fmt.Errorf("invalid timeout %d", timeout)
	}

	m, err := c.db.Multi()
	if err != nil {
		return err
	}
	defer m.Close()

	keys, err := xscan(m.DB, tp, int(count))
	if err != nil {
		return err
	} else if len(keys) == 0 {
		c.resp.writeInteger(0)
		return nil
	}

	mc := c.app.getMigrateClient(addr)

	conn := mc.Get()

	//timeout is milliseconds
	t := time.Duration(timeout) * time.Millisecond
	conn.SetConnectTimeout(t)

	if _, err = conn.Do("select", db); err != nil {
		return err
	}

	for _, key := range keys {
		data, err := xdump(m.DB, tp, key)
		if err != nil {
			return err
		}

		ttl, err := xttl(m.DB, tp, key)
		if err != nil {
			return err
		}

		conn.SetReadDeadline(time.Now().Add(t))

		//ttl is second, but restore need millisecond
		if _, err = conn.Do("restore", key, ttl*1e3, data); err != nil {
			return err
		}

		if err = xdel(m.DB, tp, key); err != nil {
			return err
		}

	}

	c.resp.writeInteger(int64(len(keys)))

	return nil
}

//XMIGRATE host port type key destination-db timeout
//will block any other write operations
//maybe only for xcodis
func xmigrateCommand(c *client) error {
	args := c.args

	if len(args) != 6 {
		return ErrCmdParams
	}

	addr := fmt.Sprintf("%s:%s", string(args[0]), string(args[1]))
	if addr == c.app.cfg.Addr {
		//same server， can not migrate
		return fmt.Errorf("migrate in same server is not allowed")
	}

	tp := string(args[2])
	key := args[3]
	db, err := ledis.StrUint64(args[4], nil)
	if err != nil {
		return err
	} else if db >= uint64(ledis.MaxDBNumber) {
		return fmt.Errorf("invalid db index %d, must < %d", db, ledis.MaxDBNumber)
	}

	timeout, err := ledis.StrInt64(args[5], nil)
	if err != nil {
		return err
	} else if timeout < 0 {
		return fmt.Errorf("invalid timeout %d", timeout)
	}

	m, err := c.db.Multi()
	if err != nil {
		return err
	}
	defer m.Close()

	data, err := xdump(m.DB, tp, key)
	if err != nil {
		return err
	} else if data == nil {
		c.resp.writeStatus(NOKEY)
		return nil
	}

	ttl, err := xttl(m.DB, tp, key)
	if err != nil {
		return err
	}

	mc := c.app.getMigrateClient(addr)

	conn := mc.Get()

	//timeout is milliseconds
	t := time.Duration(timeout) * time.Millisecond
	conn.SetConnectTimeout(t)

	if _, err = conn.Do("select", db); err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(t))

	//ttl is second, but restore need millisecond
	if _, err = conn.Do("restore", key, ttl*1e3, data); err != nil {
		return err
	}

	if err = xdel(m.DB, tp, key); err != nil {
		return err
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
	register("xmigratedb", xmigratedbCommand)
}
