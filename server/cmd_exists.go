package server

import (
	"fmt"
	"strings"

	"github.com/siddontang/ledisdb/ledis"
)

func xexists(db *ledis.DB, tp string, key []byte) (int64, error) {
	switch strings.ToUpper(tp) {
	case KVName:
		return db.Exists(key)
	case HashName:
		return db.HKeyExists(key)
	case ListName:
		return db.LKeyExists(key)
	case SetName:
		return db.SKeyExists(key)
	case ZSetName:
		return db.ZKeyExists(key)
	default:
		return 0, fmt.Errorf("invalid key type %s", tp)
	}
}

func xexistsCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}
	tp := strings.ToUpper(string(args[0]))
	key := args[1]
	if i, err := xexists(c.db, tp, key); err != nil {
		return err
	} else {
		c.resp.writeInteger(i)
	}
	return nil
}

func init() {
	register("xexists", xexistsCommand)
}
