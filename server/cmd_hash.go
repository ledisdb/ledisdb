package server

import (
	"github.com/siddontang/ledisdb/ledis"
)

func hsetCommand(req *requestContext) error {
	args := req.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	if n, err := req.db.HSet(args[0], args[1], args[2]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func hgetCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if v, err := req.db.HGet(args[0], args[1]); err != nil {
		return err
	} else {
		req.resp.writeBulk(v)
	}

	return nil
}

func hexistsCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	var n int64 = 1
	if v, err := req.db.HGet(args[0], args[1]); err != nil {
		return err
	} else {
		if v == nil {
			n = 0
		}

		req.resp.writeInteger(n)
	}
	return nil
}

func hdelCommand(req *requestContext) error {
	args := req.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := req.db.HDel(args[0], args[1:]...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func hlenCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.HLen(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func hincrbyCommand(req *requestContext) error {
	args := req.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	delta, err := ledis.StrInt64(args[2], nil)
	if err != nil {
		return err
	}

	var n int64
	if n, err = req.db.HIncrBy(args[0], args[1], delta); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}
	return nil
}

func hmsetCommand(req *requestContext) error {
	args := req.args
	if len(args) < 3 {
		return ErrCmdParams
	}

	if len(args[1:])%2 != 0 {
		return ErrCmdParams
	}

	key := args[0]

	args = args[1:]

	kvs := make([]ledis.FVPair, len(args)/2)
	for i := 0; i < len(kvs); i++ {
		kvs[i].Field = args[2*i]
		kvs[i].Value = args[2*i+1]
	}

	if err := req.db.HMset(key, kvs...); err != nil {
		return err
	} else {
		req.resp.writeStatus(OK)
	}

	return nil
}

func hmgetCommand(req *requestContext) error {
	args := req.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if v, err := req.db.HMget(args[0], args[1:]...); err != nil {
		return err
	} else {
		req.resp.writeSliceArray(v)
	}

	return nil
}

func hgetallCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.HGetAll(args[0]); err != nil {
		return err
	} else {
		req.resp.writeFVPairArray(v)
	}

	return nil
}

func hkeysCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.HKeys(args[0]); err != nil {
		return err
	} else {
		req.resp.writeSliceArray(v)
	}

	return nil
}

func hvalsCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.HValues(args[0]); err != nil {
		return err
	} else {
		req.resp.writeSliceArray(v)
	}

	return nil
}

func hclearCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.HClear(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func hmclearCommand(req *requestContext) error {
	args := req.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	if n, err := req.db.HMclear(args...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func hexpireCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := req.db.HExpire(args[0], duration); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func hexpireAtCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := req.db.HExpireAt(args[0], when); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func httlCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.HTTL(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func hpersistCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.HPersist(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func init() {
	register("hdel", hdelCommand)
	register("hexists", hexistsCommand)
	register("hget", hgetCommand)
	register("hgetall", hgetallCommand)
	register("hincrby", hincrbyCommand)
	register("hkeys", hkeysCommand)
	register("hlen", hlenCommand)
	register("hmget", hmgetCommand)
	register("hmset", hmsetCommand)
	register("hset", hsetCommand)
	register("hvals", hvalsCommand)

	//ledisdb special command

	register("hclear", hclearCommand)
	register("hmclear", hmclearCommand)
	register("hexpire", hexpireCommand)
	register("hexpireat", hexpireAtCommand)
	register("httl", httlCommand)
	register("hpersist", hpersistCommand)
}
