package http

import (
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"
	"strings"
)

func bgetCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bget")
	}
	if v, err := db.BGet([]byte(args[0])); err != nil {
		return nil, err
	} else {
		return ledis.String(v), nil
	}
}

func bdeleteCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bdelete")
	}
	if n, err := db.BDelete([]byte(args[0])); err != nil {
		return nil, err
	} else {
		return n, err
	}
}

func bsetbitCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bsetbit")
	}
	key := []byte(args[0])
	offset, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return nil, ErrValue
	}
	val, err := strconv.ParseUint(args[2], 10, 8)
	if ori, err := db.BSetBit(key, int32(offset), uint8(val)); err != nil {
		return nil, err

	} else {
		return ori, nil
	}
}

func bgetbitCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bgetbit")
	}
	key := []byte(args[0])
	offset, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return nil, ErrValue
	}

	if v, err := db.BGetBit(key, int32(offset)); err != nil {
		return nil, err
	} else {
		return v, nil
	}

	return nil, nil
}

func bmsetbitCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bmsetbit")
	}
	key := []byte(args[0])
	if len(args[1:])%2 != 0 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bmsetbit")
	} else {
		args = args[1:]
	}
	pairs := make([]ledis.BitPair, len(args)/2)
	for i := 0; i < len(pairs); i++ {
		offset, err := strconv.ParseInt(args[i*2], 10, 32)
		if err != nil {
			return nil, err
		}
		val, err := strconv.ParseUint(args[i*2+1], 10, 8)
		if err != nil {
			return nil, err
		}
		pairs[i].Pos = int32(offset)
		pairs[i].Val = uint8(val)
	}
	if place, err := db.BMSetBit(key, pairs...); err != nil {
		return nil, err
	} else {
		return place, nil
	}
}

func bcountCommand(db *ledis.DB, args ...string) (interface{}, error) {
	argCnt := len(args)
	if argCnt > 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bcount")
	}

	var err error
	var start, end int64 = 0, -1
	if argCnt > 1 {
		if start, err = strconv.ParseInt(args[1], 10, 32); err != nil {
			return nil, ErrValue
		}
	}
	if argCnt > 2 {
		if end, err = strconv.ParseInt(args[2], 10, 32); err != nil {
			return nil, ErrValue
		}
	}
	key := []byte(args[0])
	if cnt, err := db.BCount(key, int32(start), int32(end)); err != nil {
		return nil, err
	} else {
		return cnt, nil
	}
}

func boptCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bopt")
	}
	opDesc := strings.ToLower(args[0])
	dstKey := []byte(args[1])

	var srcKeys = [][]byte{}
	if len(args) >= 3 {
		srcKeys = make([][]byte, len(args[2:]))
		for i, arg := range args[2:] {
			srcKeys[i] = []byte(arg)
		}
	}

	var op uint8
	switch opDesc {
	case "and":
		op = ledis.OPand
	case "or":
		op = ledis.OPor
	case "xor":
		op = ledis.OPxor
	case "not":
		op = ledis.OPnot
	default:
		return nil, fmt.Errorf("invalid argument '%s' for 'bopt' command", opDesc)
	}
	if blen, err := db.BOperation(op, dstKey, srcKeys...); err != nil {
		return nil, err
	} else {
		return blen, nil
	}
}

func bexpireCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bexpire")
	}
	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, err
	}
	key := []byte(args[0])
	if v, err := db.BExpire(key, duration); err != nil {
		return nil, err
	} else {
		return v, err
	}
}

func bexpireatCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bexpireat")
	}
	when, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, err
	}
	key := []byte(args[0])
	if v, err := db.BExpireAt(key, when); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func bttlCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bttl")
	}
	key := []byte(args[0])
	if v, err := db.BTTL(key); err != nil {
		return nil, err
	} else {
		return v, err
	}
}

func bpersistCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "bpersist")
	}
	key := []byte(args[0])
	if n, err := db.BPersist(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func init() {
	register("bget", bgetCommand)
	register("bdelete", bdeleteCommand)
	register("bsetbit", bsetbitCommand)
	register("bgetbit", bgetbitCommand)
	register("bmsetbit", bmsetbitCommand)
	register("bcount", bcountCommand)
	register("bopt", boptCommand)
	register("bexpire", bexpireCommand)
	register("bexpireat", bexpireatCommand)
	register("bttl", bttlCommand)
	register("bpersist", bpersistCommand)
}
