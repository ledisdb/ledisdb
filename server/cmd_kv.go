package server

import (
	"github.com/siddontang/ledisdb/ledis"
)

func getCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.Get(args[0]); err != nil {
		return err
	} else {
		req.resp.writeBulk(v)
	}
	return nil
}

func setCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if err := req.db.Set(args[0], args[1]); err != nil {
		return err
	} else {
		req.resp.writeStatus(OK)
	}

	return nil
}

func getsetCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if v, err := req.db.GetSet(args[0], args[1]); err != nil {
		return err
	} else {
		req.resp.writeBulk(v)
	}

	return nil
}

func setnxCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if n, err := req.db.SetNX(args[0], args[1]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func existsCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.Exists(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func incrCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.Incr(req.args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func decrCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.Decr(req.args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func incrbyCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	delta, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if n, err := req.db.IncryBy(req.args[0], delta); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func decrbyCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	delta, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if n, err := req.db.DecrBy(req.args[0], delta); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func delCommand(req *requestContext) error {
	args := req.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	if n, err := req.db.Del(args...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func msetCommand(req *requestContext) error {
	args := req.args
	if len(args) == 0 || len(args)%2 != 0 {
		return ErrCmdParams
	}

	kvs := make([]ledis.KVPair, len(args)/2)
	for i := 0; i < len(kvs); i++ {
		kvs[i].Key = args[2*i]
		kvs[i].Value = args[2*i+1]
	}

	if err := req.db.MSet(kvs...); err != nil {
		return err
	} else {
		req.resp.writeStatus(OK)
	}

	return nil
}

// func setexCommand(req *requestContext) error {
// 	return nil
// }

func mgetCommand(req *requestContext) error {
	args := req.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	if v, err := req.db.MGet(args...); err != nil {
		return err
	} else {
		req.resp.writeSliceArray(v)
	}

	return nil
}

func expireCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := req.db.Expire(args[0], duration); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func expireAtCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := req.db.ExpireAt(args[0], when); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func ttlCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.TTL(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func persistCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.Persist(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

// func (db *DB) Expire(key []byte, duration int6
// func (db *DB) ExpireAt(key []byte, when int64)
// func (db *DB) TTL(key []byte) (int64, error)

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
	register("expire", expireCommand)
	register("expireat", expireAtCommand)
	register("ttl", ttlCommand)
	register("persist", persistCommand)
}
