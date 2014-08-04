package server

import (
	"github.com/siddontang/ledisdb/ledis"
)

func lpushCommand(req *requestContext) error {
	args := req.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := req.db.LPush(args[0], args[1:]...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func rpushCommand(req *requestContext) error {
	args := req.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := req.db.RPush(args[0], args[1:]...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func lpopCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.LPop(args[0]); err != nil {
		return err
	} else {
		req.resp.writeBulk(v)
	}

	return nil
}

func rpopCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.RPop(args[0]); err != nil {
		return err
	} else {
		req.resp.writeBulk(v)
	}

	return nil
}

func llenCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.LLen(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func lindexCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	index, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := req.db.LIndex(args[0], int32(index)); err != nil {
		return err
	} else {
		req.resp.writeBulk(v)
	}

	return nil
}

func lrangeCommand(req *requestContext) error {
	args := req.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	var start int64
	var stop int64
	var err error

	start, err = ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	stop, err = ledis.StrInt64(args[2], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := req.db.LRange(args[0], int32(start), int32(stop)); err != nil {
		return err
	} else {
		req.resp.writeSliceArray(v)
	}

	return nil
}

func lclearCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.LClear(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func lmclearCommand(req *requestContext) error {
	args := req.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	if n, err := req.db.LMclear(args...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func lexpireCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := req.db.LExpire(args[0], duration); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func lexpireAtCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := req.db.LExpireAt(args[0], when); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func lttlCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.LTTL(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func lpersistCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.LPersist(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func init() {
	register("lindex", lindexCommand)
	register("llen", llenCommand)
	register("lpop", lpopCommand)
	register("lrange", lrangeCommand)
	register("lpush", lpushCommand)
	register("rpop", rpopCommand)
	register("rpush", rpushCommand)

	//ledisdb special command

	register("lclear", lclearCommand)
	register("lmclear", lmclearCommand)
	register("lexpire", lexpireCommand)
	register("lexpireat", lexpireAtCommand)
	register("lttl", lttlCommand)
	register("lpersist", lpersistCommand)
}
