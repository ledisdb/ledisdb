package server

import (
	"strconv"

	"github.com/siddontang/ledisdb/ledis"
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

func appendCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if n, err := c.db.Append(args[0], args[1]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func getrangeCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	key := args[0]
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return err
	}

	end, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return err
	}

	if v, err := c.db.GetRange(key, start, end); err != nil {
		return err
	} else {
		c.resp.writeBulk(v)
	}

	return nil

}

func setrangeCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	key := args[0]
	offset, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return err
	}

	value := args[2]

	if n, err := c.db.SetRange(key, offset, value); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func strlenCommand(c *client) error {
	if len(c.args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.StrLen(c.args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func parseBitRange(args [][]byte) (start int, end int, err error) {
	start = 0
	end = -1
	if len(args) > 0 {
		if start, err = strconv.Atoi(string(args[0])); err != nil {
			return
		}
	}

	if len(args) == 2 {
		if end, err = strconv.Atoi(string(args[1])); err != nil {
			return
		}
	}
	return
}

func bitcountCommand(c *client) error {
	args := c.args
	if len(args) == 0 || len(args) > 3 {
		return ErrCmdParams
	}

	key := args[0]
	start, end, err := parseBitRange(args[1:])
	if err != nil {
		return err
	}

	if n, err := c.db.BitCount(key, start, end); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func bitopCommand(c *client) error {
	args := c.args
	if len(args) < 3 {
		return ErrCmdParams
	}

	op := string(args[0])
	destKey := args[1]
	srcKeys := args[2:]

	if n, err := c.db.BitOP(op, destKey, srcKeys...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func bitposCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	key := args[0]
	bit, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return err
	}
	start, end, err := parseBitRange(args[2:])
	if err != nil {
		return err
	}

	if n, err := c.db.BitPos(key, bit, start, end); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func getbitCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	key := args[0]
	offset, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return err
	}

	if n, err := c.db.GetBit(key, offset); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func setbitCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	key := args[0]
	offset, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return err
	}

	value, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return err
	}

	if n, err := c.db.SetBit(key, offset, value); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}
	return nil
}

func init() {
	register("append", appendCommand)
	register("bitcount", bitcountCommand)
	register("bitop", bitopCommand)
	register("bitpos", bitposCommand)
	register("decr", decrCommand)
	register("decrby", decrbyCommand)
	register("del", delCommand)
	register("exists", existsCommand)
	register("get", getCommand)
	register("getbit", getbitCommand)
	register("getrange", getrangeCommand)
	register("getset", getsetCommand)
	register("incr", incrCommand)
	register("incrby", incrbyCommand)
	register("mget", mgetCommand)
	register("mset", msetCommand)
	register("set", setCommand)
	register("setbit", setbitCommand)
	register("setnx", setnxCommand)
	register("setex", setexCommand)
	register("setrange", setrangeCommand)
	register("strlen", strlenCommand)
	register("expire", expireCommand)
	register("expireat", expireAtCommand)
	register("ttl", ttlCommand)
	register("persist", persistCommand)
}
