package http

import (
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"
)

func lpushCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lpush")
	}
	key := []byte(args[0])
	elems := make([][]byte, len(args[1:]))
	for i, arg := range args[1:] {
		elems[i] = []byte(arg)
	}

	if n, err := db.LPush(key, elems...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func rpushCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "rpush")
	}

	key := []byte(args[0])
	elems := make([][]byte, len(args[1:]))
	for i, arg := range args[1:] {
		elems[i] = []byte(arg)
	}
	if n, err := db.RPush(key, elems...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func lpopCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lpop")
	}

	key := []byte(args[0])

	if v, err := db.LPop(key); err != nil {
		return nil, err
	} else {
		if v == nil {
			return nil, nil
		}
		return ledis.String(v), nil
	}
}

func rpopCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "rpop")
	}
	key := []byte(args[0])

	if v, err := db.RPop(key); err != nil {
		return nil, err
	} else {
		if v == nil {
			return nil, nil
		}
		return ledis.String(v), nil
	}
}

func llenCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "llen")
	}

	key := []byte(args[0])
	if n, err := db.LLen(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func lindexCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lindex")
	}

	index, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return nil, ErrValue
	}
	key := []byte(args[0])

	if v, err := db.LIndex(key, int32(index)); err != nil {
		return nil, err
	} else {
		if v == nil {
			return nil, nil
		}
		return ledis.String(v), nil
	}
}

func lrangeCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lrange")
	}

	start, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	stop, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	key := []byte(args[0])
	if vals, err := db.LRange(key, int32(start), int32(stop)); err != nil {
		return nil, err
	} else {
		arr := make([]interface{}, len(vals))
		for i, v := range vals {
			if v == nil {
				arr[i] = nil
			} else {
				arr[i] = ledis.String(v)
			}
		}
		return arr, nil
	}
}

func lclearCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lclear")
	}

	key := []byte(args[0])
	if n, err := db.LClear(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func lmclearCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lmclear")
	}

	keys := make([][]byte, len(args))
	for i, arg := range args {
		keys[i] = []byte(arg)
	}
	if n, err := db.LMclear(keys...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func lexpireCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lexpire")
	}

	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	key := []byte(args[0])
	if v, err := db.LExpire(key, duration); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func lexpireAtCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lexpireat")
	}

	when, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	key := []byte(args[0])
	if v, err := db.LExpireAt(key, when); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func lttlCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lttl")
	}

	key := []byte(args[0])
	if v, err := db.LTTL(key); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func lpersistCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "lpersist")
	}
	key := []byte(args[0])
	if n, err := db.LPersist(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
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
