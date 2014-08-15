package server

import (
	"github.com/siddontang/ledisdb/ledis"
)

func saddCommand(req *requestContext) error {
	args := req.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := req.db.SAdd(args[0], args[1:]...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func soptGeneric(req *requestContext, optType byte) error {
	args := req.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	var v [][]byte
	var err error

	switch optType {
	case ledis.UnionType:
		v, err = req.db.SUnion(args...)
	case ledis.DiffType:
		v, err = req.db.SDiff(args...)
	case ledis.InterType:
		v, err = req.db.SInter(args...)
	}

	if err != nil {
		return err
	} else {
		req.resp.writeSliceArray(v)
	}

	return nil

}

func soptStoreGeneric(req *requestContext, optType byte) error {
	args := req.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	var n int64
	var err error

	switch optType {
	case ledis.UnionType:
		n, err = req.db.SUnionStore(args[0], args[1:]...)
	case ledis.DiffType:
		n, err = req.db.SDiffStore(args[0], args[1:]...)
	case ledis.InterType:
		n, err = req.db.SInterStore(args[0], args[1:]...)
	}

	if err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func scardCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.SCard(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func sdiffCommand(req *requestContext) error {
	return soptGeneric(req, ledis.DiffType)
}

func sdiffstoreCommand(req *requestContext) error {
	return soptStoreGeneric(req, ledis.DiffType)
}

func sinterCommand(req *requestContext) error {
	return soptGeneric(req, ledis.InterType)

}

func sinterstoreCommand(req *requestContext) error {
	return soptStoreGeneric(req, ledis.InterType)
}

func sismemberCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if n, err := req.db.SIsMember(args[0], args[1]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func smembersCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.SMembers(args[0]); err != nil {
		return err
	} else {
		req.resp.writeSliceArray(v)
	}

	return nil

}

func sremCommand(req *requestContext) error {
	args := req.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := req.db.SRem(args[0], args[1:]...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil

}

func sunionCommand(req *requestContext) error {
	return soptGeneric(req, ledis.UnionType)
}

func sunionstoreCommand(req *requestContext) error {
	return soptStoreGeneric(req, ledis.UnionType)
}

func sclearCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.SClear(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func smclearCommand(req *requestContext) error {
	args := req.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	if n, err := req.db.SMclear(args...); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func sexpireCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := req.db.SExpire(args[0], duration); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func sexpireAtCommand(req *requestContext) error {
	args := req.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := req.db.SExpireAt(args[0], when); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil
}

func sttlCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := req.db.STTL(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(v)
	}

	return nil

}

func spersistCommand(req *requestContext) error {
	args := req.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := req.db.SPersist(args[0]); err != nil {
		return err
	} else {
		req.resp.writeInteger(n)
	}

	return nil
}

func init() {
	register("sadd", saddCommand)
	register("scard", scardCommand)
	register("sdiff", sdiffCommand)
	register("sdiffstore", sdiffstoreCommand)
	register("sinter", sinterCommand)
	register("sinterstore", sinterstoreCommand)
	register("sismember", sismemberCommand)
	register("smembers", smembersCommand)
	register("srem", sremCommand)
	register("sunion", sunionCommand)
	register("sunionstore", sunionstoreCommand)
	register("sclear", sclearCommand)
	register("smclear", smclearCommand)
	register("sexpire", sexpireCommand)
	register("sexpireat", sexpireAtCommand)
	register("sttl", sttlCommand)
	register("spersist", spersistCommand)
}
