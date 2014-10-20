package server

import (
	"github.com/siddontang/ledisdb/ledis"
)

func saddCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.db.SAdd(args[0], args[1:]...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func soptGeneric(c *client, optType byte) error {
	args := c.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	var v [][]byte
	var err error

	switch optType {
	case ledis.UnionType:
		v, err = c.db.SUnion(args...)
	case ledis.DiffType:
		v, err = c.db.SDiff(args...)
	case ledis.InterType:
		v, err = c.db.SInter(args...)
	}

	if err != nil {
		return err
	} else {
		c.resp.writeSliceArray(v)
	}

	return nil

}

func soptStoreGeneric(c *client, optType byte) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	var n int64
	var err error

	switch optType {
	case ledis.UnionType:
		n, err = c.db.SUnionStore(args[0], args[1:]...)
	case ledis.DiffType:
		n, err = c.db.SDiffStore(args[0], args[1:]...)
	case ledis.InterType:
		n, err = c.db.SInterStore(args[0], args[1:]...)
	}

	if err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func scardCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.SCard(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func sdiffCommand(c *client) error {
	return soptGeneric(c, ledis.DiffType)
}

func sdiffstoreCommand(c *client) error {
	return soptStoreGeneric(c, ledis.DiffType)
}

func sinterCommand(c *client) error {
	return soptGeneric(c, ledis.InterType)

}

func sinterstoreCommand(c *client) error {
	return soptStoreGeneric(c, ledis.InterType)
}

func sismemberCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if n, err := c.db.SIsMember(args[0], args[1]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func smembersCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.SMembers(args[0]); err != nil {
		return err
	} else {
		c.resp.writeSliceArray(v)
	}

	return nil

}

func sremCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.db.SRem(args[0], args[1:]...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil

}

func sunionCommand(c *client) error {
	return soptGeneric(c, ledis.UnionType)
}

func sunionstoreCommand(c *client) error {
	return soptStoreGeneric(c, ledis.UnionType)
}

func sclearCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.SClear(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func smclearCommand(c *client) error {
	args := c.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	if n, err := c.db.SMclear(args...); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func sexpireCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.SExpire(args[0], duration); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func sexpireAtCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	if v, err := c.db.SExpireAt(args[0], when); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func sttlCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.STTL(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil

}

func spersistCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.SPersist(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func sxscanCommand(c *client) error {
	return xscanGeneric(c, c.db.SScan)
}

func sxrevscanCommand(c *client) error {
	return xscanGeneric(c, c.db.SRevScan)
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
	register("sxscan", sxscanCommand)
	register("sxrevscan", sxrevscanCommand)
}
