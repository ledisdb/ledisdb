package http

import (
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"
)

func hsetCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hset")
	}

	key := []byte(args[0])
	field := []byte(args[1])
	value := []byte(args[2])
	if n, err := db.HSet(key, field, value); err != nil {
		return nil, err
	} else {
		return n, err
	}
}

func hgetCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hget")
	}

	key := []byte(args[0])
	field := []byte(args[1])

	if v, err := db.HGet(key, field); err != nil {
		return nil, err
	} else {
		if v == nil {
			return nil, nil
		}
		return ledis.String(v), nil
	}
}

func hexistsCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hexists")
	}
	key := []byte(args[0])
	field := []byte(args[1])

	var n int64 = 1
	if v, err := db.HGet(key, field); err != nil {
		return nil, err
	} else {
		if v == nil {
			n = 0
		}
		return n, nil
	}
}

func hdelCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hdel")
	}
	key := []byte(args[0])
	fields := make([][]byte, len(args[1:]))
	for i, arg := range args[1:] {
		fields[i] = []byte(arg)
	}

	if n, err := db.HDel(key, fields...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func hlenCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hlen")
	}
	key := []byte(args[0])
	if n, err := db.HLen(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func hincrbyCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hincrby")
	}
	key := []byte(args[0])
	field := []byte(args[1])
	delta, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	var n int64
	if n, err = db.HIncrBy(key, field, delta); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func hmsetCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hmset")
	}

	if len(args[1:])%2 != 0 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hmset")
	}
	key := []byte(args[0])
	args = args[1:]
	kvs := make([]ledis.FVPair, len(args)/2)
	for i := 0; i < len(kvs); i++ {
		kvs[i].Field = []byte(args[2*i])
		kvs[i].Value = []byte(args[2*i+1])
	}
	if err := db.HMset(key, kvs...); err != nil {
		return nil, err
	} else {
		return []interface{}{true, MSG_OK}, nil
	}
}

func hmgetCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hmget")
	}
	key := []byte(args[0])
	fields := make([][]byte, len(args[1:]))
	for i, arg := range args[1:] {
		fields[i] = []byte(arg)
	}
	if vals, err := db.HMget(key, fields...); err != nil {
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

func hgetallCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hgetall")
	}
	key := []byte(args[0])
	if fvs, err := db.HGetAll(key); err != nil {
		return nil, err
	} else {
		var m = make(map[string]string)
		for _, fv := range fvs {
			m[ledis.String(fv.Field)] = ledis.String(fv.Value)
		}
		return m, nil
	}
}

func hkeysCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hkeys")
	}
	key := []byte(args[0])
	if fields, err := db.HKeys(key); err != nil {
		return nil, err
	} else {
		arr := make([]string, len(fields))
		for i, f := range fields {
			arr[i] = ledis.String(f)
		}
		return arr, nil
	}
}

func hvalsCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hvals")
	}
	key := []byte(args[0])
	if vals, err := db.HValues(key); err != nil {
		return nil, err
	} else {
		var arr = make([]string, len(vals))
		for i, v := range vals {
			arr[i] = ledis.String(v)
		}
		return arr, nil
	}
}

func hclearCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hclear")
	}
	key := []byte(args[0])
	if n, err := db.HClear(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func hmclearCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hmclear")
	}
	keys := make([][]byte, len(args))
	for i, arg := range args {
		keys[i] = []byte(arg)
	}

	if n, err := db.HMclear(keys...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func hexpireCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hexpire")
	}
	key := []byte(args[0])
	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}
	if v, err := db.HExpire(key, duration); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func hexpireAtCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hexpireat")
	}
	key := []byte(args[0])

	when, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	if v, err := db.HExpireAt(key, when); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func httlCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "httl")
	}

	key := []byte(args[0])
	if v, err := db.HTTL(key); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func hpersistCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "hpersist")
	}
	key := []byte(args[0])
	if n, err := db.HPersist(key); err != nil {
		return nil, err
	} else {
		return n, err
	}
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
