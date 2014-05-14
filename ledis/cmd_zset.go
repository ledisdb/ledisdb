package ledis

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

//for simple implementation, we only support int64 score

const (
	MinScore int64 = -1<<63 + 1
	MaxScore int64 = 1<<63 - 1
)

var errScoreOverflow = errors.New("zset score overflow")

func zaddCommand(c *client) error {
	args := c.args
	if len(args) < 3 {
		return ErrCmdParams
	}

	key := args[0]
	if len(args[1:])%2 != 0 {
		return ErrCmdParams
	}

	args = args[1:]
	params := make([]interface{}, len(args))
	for i := 0; i < len(params); i += 2 {
		score, err := StrInt64(args[i], nil)
		if err != nil {
			return err
		}

		params[i] = score
		params[i+1] = args[i+1]
	}

	if n, err := c.app.zset_add(key, params); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zcardCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.app.zset_card(args[0]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zscoreCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if v, err := c.app.zset_score(args[0], args[1]); err != nil {
		return err
	} else {
		c.writeBulk(v)
	}

	return nil
}

func zremCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.app.zset_rem(args[0], args[1:]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zincrbyCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	key := args[0]

	delta, err := StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := c.app.zset_incrby(key, delta, args[2]); err != nil {
		return err
	} else {
		c.writeBulk(v)
	}

	return nil
}

func zparseScoreRange(minBuf []byte, maxBuf []byte) (min int64, max int64, err error) {
	if strings.ToLower(String(minBuf)) == "-inf" {
		min = math.MinInt64
	} else {
		var lopen bool = false
		if minBuf[0] == '(' {
			lopen = true
			minBuf = minBuf[1:]
		}

		if len(minBuf) == 0 {
			err = ErrCmdParams
			return
		}

		min, err = StrInt64(minBuf, nil)
		if err != nil {
			return
		}

		if min <= MinScore || min >= MaxScore {
			err = errScoreOverflow
			return
		}

		if lopen {
			min++
		}
	}

	if strings.ToLower(String(maxBuf)) == "+inf" {
		max = math.MaxInt64
	} else {
		var ropen = false
		if maxBuf[0] == '(' {
			ropen = true
			maxBuf = maxBuf[1:]
		}

		if len(maxBuf) == 0 {
			err = ErrCmdParams
			return
		}

		max, err = StrInt64(maxBuf, nil)
		if err != nil {
			return
		}

		if max <= MinScore || max >= MaxScore {
			err = errScoreOverflow
			return
		}

		if ropen {
			max--
		}
	}

	return
}

func zcountCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	min, max, err := zparseScoreRange(args[1], args[2])
	if err != nil {
		return err
	}

	if min > max {
		c.writeInteger(0)
		return nil
	}

	if n, err := c.app.zset_count(args[0], min, max); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zrankCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if n, err := c.app.zset_rank(args[0], args[1], false); err != nil {
		return err
	} else if n == -1 {
		c.writeBulk(nil)
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zrevrankCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	if n, err := c.app.zset_rank(args[0], args[1], true); err != nil {
		return err
	} else if n == -1 {
		c.writeBulk(nil)
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zremrangebyrankCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	key := args[0]

	offset, limit, err := zparseRange(c, key, args[1], args[2])
	if err != nil {
		return err
	}

	if offset < 0 {
		c.writeInteger(0)
		return nil
	}

	if n, err := c.app.zset_remRange(key, MinScore, MaxScore, offset, limit); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zremrangebyscoreCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	key := args[0]
	min, max, err := zparseScoreRange(args[1], args[2])
	if err != nil {
		return err
	}

	if n, err := c.app.zset_remRange(key, min, max, 0, -1); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zparseRange(c *client, key []byte, startBuf []byte, stopBuf []byte) (offset int, limit int, err error) {
	var start int
	var stop int
	if start, err = strconv.Atoi(String(startBuf)); err != nil {
		return
	}

	if stop, err = strconv.Atoi(String(stopBuf)); err != nil {
		return
	}

	if start < 0 || stop < 0 {
		//refer redis implementation
		var size int64
		size, err = c.app.zset_card(key)
		if err != nil {
			return
		}

		llen := int(size)

		if start < 0 {
			start = llen + start
		}
		if stop < 0 {
			stop = llen + stop
		}

		if start < 0 {
			start = 0
		}

		if start >= llen {
			offset = -1
			return
		}
	}

	if start > stop {
		offset = -1
		return
	}

	offset = start
	limit = (stop - start) + 1
	return
}

func zrangeGeneric(c *client, reverse bool) error {
	args := c.args
	if len(args) < 3 {
		return ErrCmdParams
	}

	key := args[0]

	offset, limit, err := zparseRange(c, key, args[1], args[2])
	if err != nil {
		return err
	}

	if offset < 0 {
		c.writeArray([]interface{}{})
		return nil
	}

	args = args[3:]
	var withScores bool = false

	if len(args) > 0 && strings.ToLower(String(args[0])) == "withscores" {
		withScores = true
	}

	if v, err := c.app.zset_range(key, MinScore, MaxScore, withScores, offset, limit, reverse); err != nil {
		return err
	} else {
		c.writeArray(v)
	}
	return nil
}

func zrangeCommand(c *client) error {
	return zrangeGeneric(c, false)
}

func zrevrangeCommand(c *client) error {
	return zrangeGeneric(c, true)
}

func zrangebyscoreGeneric(c *client, reverse bool) error {
	args := c.args
	if len(args) < 3 {
		return ErrCmdParams
	}

	key := args[0]
	min, max, err := zparseScoreRange(args[1], args[2])
	if err != nil {
		return err
	}

	args = args[3:]

	var withScores bool = false

	if len(args) > 0 && strings.ToLower(String(args[0])) == "withscores" {
		withScores = true
		args = args[1:]
	}

	var offset int = 0
	var limit int = -1

	if len(args) > 0 {
		if len(args) != 3 {
			return ErrCmdParams
		}

		if strings.ToLower(String(args[0])) != "limit" {
			return ErrCmdParams
		}

		if offset, err = strconv.Atoi(String(args[1])); err != nil {
			return ErrCmdParams
		}

		if limit, err = strconv.Atoi(String(args[2])); err != nil {
			return ErrCmdParams
		}
	}

	if offset < 0 {
		//for redis, if offset < 0, a empty will return
		//so here we directly return a empty array
		c.writeArray([]interface{}{})
		return nil
	}

	if v, err := c.app.zset_range(key, min, max, withScores, offset, limit, reverse); err != nil {
		return err
	} else {
		c.writeArray(v)
	}

	return nil
}

func zrangebyscoreCommand(c *client) error {
	return zrangebyscoreGeneric(c, false)
}

func zrevrangebyscoreCommand(c *client) error {
	return zrangebyscoreGeneric(c, true)
}

func zclearCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.app.zset_clear(args[0]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func init() {
	register("zadd", zaddCommand)
	register("zcard", zcardCommand)
	register("zcount", zcountCommand)
	register("zincrby", zincrbyCommand)
	register("zrange", zrangeCommand)
	register("zrangebyscore", zrangebyscoreCommand)
	register("zrank", zrankCommand)
	register("zrem", zremCommand)
	register("zremrangebyrank", zremrangebyrankCommand)
	register("zremrangebyscore", zremrangebyscoreCommand)
	register("zrevrange", zrevrangeCommand)
	register("zrevrank", zrevrankCommand)
	register("zrevrangebyscore", zrevrangebyscoreCommand)
	register("zscore", zscoreCommand)

	//ledisdb special command
	register("zclear", zclearCommand)

}
