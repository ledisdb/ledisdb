package http

import (
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"
)

func getCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "get")
	}
	key := []byte(args[0])
	if v, err := db.Get(key); err != nil {
		return nil, err
	} else {
		if v == nil {
			return nil, nil
		}
		return ledis.String(v), nil
	}
}

func setCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "set")
	}

	key := []byte(args[0])
	val := []byte(args[1])
	if err := db.Set(key, val); err != nil {
		return nil, err
	} else {
		return []interface{}{true, MSG_OK}, nil
	}

}

func getsetCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "getset")
	}

	key := []byte(args[0])
	val := []byte(args[1])
	if v, err := db.GetSet(key, val); err != nil {
		return nil, err
	} else {
		if v == nil {
			return nil, nil
		}
		return ledis.String(v), nil
	}
}

func setnxCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "setnx")
	}

	key := []byte(args[0])
	val := []byte(args[1])
	if n, err := db.SetNX(key, val); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func existsCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "exists")
	}

	if n, err := db.Exists([]byte(args[0])); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func incrCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "incr")
	}

	key := []byte(args[0])
	if n, err := db.Incr(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func decrCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "decr")
	}

	key := []byte(args[0])
	if n, err := db.Decr(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func incrbyCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "incrby")
	}

	delta, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	key := []byte(args[0])

	if n, err := db.IncryBy(key, delta); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func decrbyCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "decrby")
	}

	delta, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	key := []byte(args[0])

	if n, err := db.DecrBy(key, delta); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func delCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "del")
	}

	keys := make([][]byte, len(args))
	if n, err := db.Del(keys...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func msetCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) == 0 || len(args)%2 != 0 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "mset")
	}

	kvs := make([]ledis.KVPair, len(args)/2)
	for i := 0; i < len(kvs); i++ {
		kvs[i].Key = []byte(args[2*i])
		kvs[i].Value = []byte(args[2*i+1])
	}

	if err := db.MSet(kvs...); err != nil {
		return nil, err
	} else {
		return []interface{}{true, MSG_OK}, nil
	}
}

func mgetCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "mget")
	}

	keys := make([][]byte, len(args))
	for i, arg := range args {
		keys[i] = []byte(arg)
	}
	if vals, err := db.MGet(keys...); err != nil {
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

func expireCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "expire")
	}

	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}
	key := []byte(args[0])
	if v, err := db.Expire(key, duration); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func expireAtCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "expireat")
	}

	when, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}
	key := []byte(args[0])
	if v, err := db.ExpireAt(key, when); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func ttlCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "ttl")
	}
	key := []byte(args[0])

	if v, err := db.TTL(key); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func persistCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "persist")
	}
	key := []byte(args[0])

	if n, err := db.Persist(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
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
	register("expire", expireCommand)
	register("expireat", expireAtCommand)
	register("ttl", ttlCommand)
	register("persist", persistCommand)
}
