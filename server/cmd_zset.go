package server

import (
	"errors"
	"github.com/siddontang/ledisdb/ledis"
	"math"
	"strconv"
	"strings"
)

//for simple implementation, we only support int64 score

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

	params := make([]ledis.ScorePair, len(args)/2)
	for i := 0; i < len(params); i++ {
		score, err := ledis.StrInt64(args[2*i], nil)
		if err != nil {
			return err
		}

		params[i].Score = score
		params[i].Member = args[2*i+1]
	}

	if n, err := c.db.ZAdd(key, params...); err != nil {
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

	if n, err := c.db.ZCard(args[0]); err != nil {
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

	if s, err := c.db.ZScore(args[0], args[1]); err != nil {
		if err == ledis.ErrScoreMiss {
			c.writeBulk(nil)
		} else {
			return err
		}
	} else {
		c.writeBulk(ledis.StrPutInt64(s))
	}

	return nil
}

func zremCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	if n, err := c.db.ZRem(args[0], args[1:]...); err != nil {
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

	delta, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := c.db.ZIncrBy(key, delta, args[2]); err != nil {
		return err
	} else {
		c.writeBulk(ledis.StrPutInt64(v))
	}

	return nil
}

func zparseScoreRange(minBuf []byte, maxBuf []byte) (min int64, max int64, err error) {
	if strings.ToLower(ledis.String(minBuf)) == "-inf" {
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

		min, err = ledis.StrInt64(minBuf, nil)
		if err != nil {
			return
		}

		if min <= ledis.MinScore || min >= ledis.MaxScore {
			err = errScoreOverflow
			return
		}

		if lopen {
			min++
		}
	}

	if strings.ToLower(ledis.String(maxBuf)) == "+inf" {
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

		max, err = ledis.StrInt64(maxBuf, nil)
		if err != nil {
			return
		}

		if max <= ledis.MinScore || max >= ledis.MaxScore {
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

	if n, err := c.db.ZCount(args[0], min, max); err != nil {
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

	if n, err := c.db.ZRank(args[0], args[1]); err != nil {
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

	if n, err := c.db.ZRevRank(args[0], args[1]); err != nil {
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

	start, stop, err := zparseRange(c, args[1], args[2])
	if err != nil {
		return err
	}

	if n, err := c.db.ZRemRangeByRank(key, start, stop); err != nil {
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

	if n, err := c.db.ZRemRangeByScore(key, min, max); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zparseRange(c *client, a1 []byte, a2 []byte) (start int, stop int, err error) {
	if start, err = strconv.Atoi(ledis.String(a1)); err != nil {
		return
	}

	if stop, err = strconv.Atoi(ledis.String(a2)); err != nil {
		return
	}

	return
}

func zrangeGeneric(c *client, reverse bool) error {
	args := c.args
	if len(args) < 3 {
		return ErrCmdParams
	}

	key := args[0]

	start, stop, err := zparseRange(c, args[1], args[2])
	if err != nil {
		return err
	}

	args = args[3:]
	var withScores bool = false

	if len(args) > 0 && strings.ToLower(ledis.String(args[0])) == "withscores" {
		withScores = true
	}

	if datas, err := c.db.ZRangeGeneric(key, start, stop, reverse); err != nil {
		return err
	} else {
		c.writeScorePairArray(datas, withScores)
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

	if len(args) > 0 && strings.ToLower(ledis.String(args[0])) == "withscores" {
		withScores = true
		args = args[1:]
	}

	var offset int = 0
	var count int = -1

	if len(args) > 0 {
		if len(args) != 3 {
			return ErrCmdParams
		}

		if strings.ToLower(ledis.String(args[0])) != "limit" {
			return ErrCmdParams
		}

		if offset, err = strconv.Atoi(ledis.String(args[1])); err != nil {
			return ErrCmdParams
		}

		if count, err = strconv.Atoi(ledis.String(args[2])); err != nil {
			return ErrCmdParams
		}
	}

	if offset < 0 {
		//for ledis, if offset < 0, a empty will return
		//so here we directly return a empty array
		c.writeArray([]interface{}{})
		return nil
	}

	if datas, err := c.db.ZRangeByScoreGeneric(key, min, max, offset, count, reverse); err != nil {
		return err
	} else {
		c.writeScorePairArray(datas, withScores)
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

	if n, err := c.db.ZClear(args[0]); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zmclearCommand(c *client) error {
	args := c.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	if n, err := c.db.ZMclear(args...); err != nil {
		return err
	} else {
		c.writeInteger(n)
	}

	return nil
}

func zexpireCommand(c *client) error {
	args := c.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := c.db.ZExpire(args[0], duration); err != nil {
		return err
	} else {
		c.writeInteger(v)
	}

	return nil
}

func zexpireAtCommand(c *client) error {
	args := c.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return err
	}

	if v, err := c.db.ZExpireAt(args[0], when); err != nil {
		return err
	} else {
		c.writeInteger(v)
	}

	return nil
}

func zttlCommand(c *client) error {
	args := c.args
	if len(args) == 0 {
		return ErrCmdParams
	}

	if v, err := c.db.ZTTL(args[0]); err != nil {
		return err
	} else {
		c.writeInteger(v)
	}

	return nil
}

func zpersistCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.ZPersist(args[0]); err != nil {
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
	register("zmclear", zmclearCommand)
	register("zexpire", zexpireCommand)
	register("zexpireat", zexpireAtCommand)
	register("zttl", zttlCommand)
	register("zpersist", zpersistCommand)
}
