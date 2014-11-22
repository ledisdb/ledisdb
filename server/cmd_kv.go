package server

import (
	"github.com/siddontang/go/hack"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"
	"strings"
)

// func getCommand(c *client) error {
// 	args := c.args
// 	if len(args) != 1 {
// 		return ErrCmdParams
// 	}

// 	if v, err := c.db.Get(args[0]); err != nil {
// 		return err
// 	} else {
// 		c.resp.writeBulk(v)
// 	}
// 	return nil
// }

func getCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.GetSlice(args[0]); err != nil {
		return err
	} else {
		if v == nil {
			c.resp.writeBulk(nil)
		} else {
			c.resp.writeBulk(v.Data())
			v.Free()
		}
	}
	return nil
}

func setCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if err := c.db.Set(args[0], args[1]); err != nil {
		return err
	} else {
		c.resp.writeStatus(OK)
	}

	return nil
}

func getsetCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if v, err := c.db.GetSet(args[0], args[1]); err != nil {
		return err
	} else {
		c.resp.writeBulk(v)
	}

	return nil
}

func setnxCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if n, err := c.db.SetNX(args[0], args[1]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func setexCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	sec, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if err := c.db.SetEX(args[0], sec, args[2]); err != nil {
		return err
	} else {
		c.resp.writeStatus(OK)
	}

	return nil
}

func existsCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.Exists(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func incrCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.Incr(c.args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func decrCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.Decr(c.args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func incrbyCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	delta, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if n, err := c.db.IncrBy(c.args[0], delta); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func decrbyCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	delta, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if n, err := c.db.DecrBy(c.args[0], delta); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func delCommand(c *client) error {
	args := c.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	if n, err := c.db.Del(args...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func msetCommand(c *client) error {
	args := c.args
	if len(args) == 0 || len(args)%2 != 0 {
		return ErrCmdParams
	}

	kvs := make([]ledis.KVPair, len(args)/2)
	for i := 0; i < len(kvs); i++ {
		kvs[i].Key = args[2*i]
		kvs[i].Value = args[2*i+1]
	}

	if err := c.db.MSet(kvs...); err != nil {
		return err
	} else {
		c.resp.writeStatus(OK)
	}

	return nil
}

// func setexCommand(c *client) error {
// 	return nil
// }

func mgetCommand(c *client) error {
	args := c.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	if v, err := c.db.MGet(args...); err != nil {
		return err
	} else {
		c.resp.writeSliceArray(v)
	}

	return nil
}

func expireCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.Expire(args[0], duration); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func expireAtCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.ExpireAt(args[0], when); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func ttlCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.TTL(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func persistCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.Persist(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func parseScanArgs(c *client) (key []byte, match string, count int, err error) {
	args := c.args
	count = 10

	switch len(args) {
	case 0:
		key = nil
		return
	case 1, 3, 5:
		key = args[0]
		break
	default:
		err = ErrCmdParams
		return
	}

	if len(args) == 3 {
		switch strings.ToLower(hack.String(args[1])) {
		case "match":
			match = hack.String(args[2])
		case "count":
			count, err = strconv.Atoi(hack.String(args[2]))
		default:
			err = ErrCmdParams
			return
		}
	} else if len(args) == 5 {
		if strings.ToLower(hack.String(args[1])) != "match" {
			err = ErrCmdParams
			return
		} else if strings.ToLower(hack.String(args[3])) != "count" {
			err = ErrCmdParams
			return
		}

		match = hack.String(args[2])
		count, err = strconv.Atoi(hack.String(args[4]))
	}

	if count <= 0 {
		err = ErrCmdParams
	}

	return
}

func xscanGeneric(c *client,
	f func(key []byte, count int, inclusive bool, match string) ([][]byte, error)) error {
	key, match, count, err := parseScanArgs(c)
	if err != nil {
		return err
	}

	if ay, err := f(key, count, false, match); err != nil {
		return err
	} else {
		data := make([]interface{}, 2)
		if len(ay) < count {
			data[0] = []byte("")
		} else {
			data[0] = ay[len(ay)-1]
		}
		data[1] = ay
		c.resp.writeArray(data)
	}
	return nil
}

func xscanCommand(c *client) error {
	return xscanGeneric(c, c.db.Scan)
}

func xrevscanCommand(c *client) error {
	return xscanGeneric(c, c.db.RevScan)
}

func init() {
	register("decr", decrCommand)
	register("decrby", decrbyCommand)
	register("del", delCommand)
	register("exists", existsCommand)
	register("get", getCommand)
	register("getset", getsetCommand)
	register("incr", incrCommand)
	register("incrby", incrbyCommand)
	register("mget", mgetCommand)
	register("mset", msetCommand)
	register("set", setCommand)
	register("setnx", setnxCommand)
	register("setex", setexCommand)
	register("expire", expireCommand)
	register("expireat", expireAtCommand)
	register("ttl", ttlCommand)
	register("persist", persistCommand)
	register("xscan", xscanCommand)
	register("xrevscan", xrevscanCommand)
}
