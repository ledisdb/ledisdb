package server

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/siddontang/go/hack"
	goledis "github.com/siddontang/ledisdb/client/goledis"
	"github.com/siddontang/ledisdb/ledis"
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
	case KVName:
		data, err = db.Dump(key)
	case HashName:
		data, err = db.HDump(key)
	case ListName:
		data, err = db.LDump(key)
	case SetName:
		data, err = db.SDump(key)
	case ZSetName:
		data, err = db.ZDump(key)
	default:
		err = fmt.Errorf("invalid key type %s", tp)
	}
	return data, err
}

func xdel(db *ledis.DB, tp string, key []byte) error {
	var err error
	switch strings.ToUpper(tp) {
	case KVName:
		_, err = db.Del(key)
	case HashName:
		_, err = db.HClear(key)
	case ListName:
		_, err = db.LClear(key)
	case SetName:
		_, err = db.SClear(key)
	case ZSetName:
		_, err = db.ZClear(key)
	default:
		err = fmt.Errorf("invalid key type %s", tp)
	}
	return err
}

func xttl(db *ledis.DB, tp string, key []byte) (int64, error) {
	switch strings.ToUpper(tp) {
	case KVName:
		return db.TTL(key)
	case HashName:
		return db.HTTL(key)
	case ListName:
		return db.LTTL(key)
	case SetName:
		return db.STTL(key)
	case ZSetName:
		return db.ZTTL(key)
	default:
		return 0, fmt.Errorf("invalid key type %s", tp)
	}
}

func xscan(db *ledis.DB, tp string, count int) ([][]byte, error) {
	switch strings.ToUpper(tp) {
	case KVName:
		return db.Scan(KV, nil, count, false, "")
	case HashName:
		return db.Scan(HASH, nil, count, false, "")
	case ListName:
		return db.Scan(LIST, nil, count, false, "")
	case SetName:
		return db.Scan(SET, nil, count, false, "")
	case ZSetName:
		return db.Scan(ZSET, nil, count, false, "")
	default:
		return nil, fmt.Errorf("invalid key type %s", tp)
	}
}

func xdumpCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	tp := strings.ToUpper(string(args[0]))
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

type migrateKeyLocker struct {
	m sync.Mutex

	locks map[string]struct{}
}

func (m *migrateKeyLocker) Lock(key []byte) bool {
	m.m.Lock()
	defer m.m.Unlock()

	k := hack.String(key)
	_, ok := m.locks[k]
	if ok {
		return false
	}
	m.locks[k] = struct{}{}
	return true
}

func (m *migrateKeyLocker) Unlock(key []byte) {
	m.m.Lock()
	defer m.m.Unlock()

	delete(m.locks, hack.String(key))
}

func newMigrateKeyLocker() *migrateKeyLocker {
	m := new(migrateKeyLocker)

	m.locks = make(map[string]struct{})

	return m
}

func (a *App) newMigrateKeyLockers() {
	a.migrateKeyLockers = make(map[string]*migrateKeyLocker)

	a.migrateKeyLockers[KVName] = newMigrateKeyLocker()
	a.migrateKeyLockers[HashName] = newMigrateKeyLocker()
	a.migrateKeyLockers[ListName] = newMigrateKeyLocker()
	a.migrateKeyLockers[SetName] = newMigrateKeyLocker()
	a.migrateKeyLockers[ZSetName] = newMigrateKeyLocker()
}

func (a *App) migrateKeyLock(tp string, key []byte) bool {
	l, ok := a.migrateKeyLockers[strings.ToUpper(tp)]
	if !ok {
		return false
	}

	return l.Lock(key)
}

func (a *App) migrateKeyUnlock(tp string, key []byte) {
	l, ok := a.migrateKeyLockers[strings.ToUpper(tp)]
	if !ok {
		return
	}

	l.Unlock(key)
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

	tp := strings.ToUpper(string(args[2]))

	count, err := ledis.StrInt64(args[3], nil)
	if err != nil {
		return err
	} else if count <= 0 {
		count = 10
	}

	db, err := ledis.StrUint64(args[4], nil)
	if err != nil {
		return err
	} else if db >= uint64(c.app.cfg.Databases) {
		return fmt.Errorf("invalid db index %d, must < %d", db, c.app.cfg.Databases)
	}

	timeout, err := ledis.StrInt64(args[5], nil)
	if err != nil {
		return err
	} else if timeout < 0 {
		return fmt.Errorf("invalid timeout %d", timeout)
	}

	keys, err := xscan(c.db, tp, int(count))
	if err != nil {
		return err
	} else if len(keys) == 0 {
		c.resp.writeInteger(0)
		return nil
	}

	mc := c.app.getMigrateClient(addr)

	conn, err := mc.Get()
	if err != nil {
		return err
	}
	defer conn.Close()

	//timeout is milliseconds
	t := time.Duration(timeout) * time.Millisecond

	if _, err = conn.Do("select", db); err != nil {
		return err
	}

	migrateNum := int64(0)
	for _, key := range keys {
		if !c.app.migrateKeyLock(tp, key) {
			// other may also migrate this key, skip it
			continue
		}

		defer c.app.migrateKeyUnlock(tp, key)

		data, err := xdump(c.db, tp, key)
		if err != nil {
			return err
		} else if data == nil {
			// no key now, skip it
			continue
		}

		ttl, err := xttl(c.db, tp, key)
		if err != nil {
			return err
		}

		conn.SetReadDeadline(time.Now().Add(t))

		//ttl is second, but restore need millisecond
		if _, err = conn.Do("restore", key, ttl*1e3, data); err != nil {
			return err
		}

		if err = xdel(c.db, tp, key); err != nil {
			return err
		}

		migrateNum++
	}

	c.resp.writeInteger(migrateNum)

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

	tp := strings.ToUpper(string(args[2]))
	key := args[3]
	db, err := ledis.StrUint64(args[4], nil)
	if err != nil {
		return err
	} else if db >= uint64(c.app.cfg.Databases) {
		return fmt.Errorf("invalid db index %d, must < %d", db, c.app.cfg.Databases)
	}

	timeout, err := ledis.StrInt64(args[5], nil)
	if err != nil {
		return err
	} else if timeout < 0 {
		return fmt.Errorf("invalid timeout %d", timeout)
	}

	if !c.app.migrateKeyLock(tp, key) {
		// other may also migrate this key, skip it
		return fmt.Errorf("%s %s is in migrating yet", tp, key)
	}

	defer c.app.migrateKeyUnlock(tp, key)

	data, err := xdump(c.db, tp, key)
	if err != nil {
		return err
	} else if data == nil {
		c.resp.writeStatus(NOKEY)
		return nil
	}

	ttl, err := xttl(c.db, tp, key)
	if err != nil {
		return err
	}

	mc := c.app.getMigrateClient(addr)

	conn, err := mc.Get()
	if err != nil {
		return err
	}
	defer conn.Close()

	//timeout is milliseconds
	t := time.Duration(timeout) * time.Millisecond

	if _, err = conn.Do("select", db); err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(t))

	//ttl is second, but restore need millisecond
	if _, err = conn.Do("restore", key, ttl*1e3, data); err != nil {
		return err
	}

	if err = xdel(c.db, tp, key); err != nil {
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
