package server

import (
	"errors"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/store"
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
	if len(args[1:])&1 != 0 {
		return ErrCmdParams
	}

	args = args[1:]

	params := make([]ledis.ScorePair, len(args)>>1)
	for i := 0; i < len(params); i++ {
		score, err := ledis.StrInt64(args[2*i], nil)
		if err != nil {
			return ErrValue
		}

		params[i].Score = score
		params[i].Member = args[2*i+1]
	}

	n, err := c.db.ZAdd(key, params...)

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
}

func zcardCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if n, err := c.db.ZCard(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
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
			c.resp.writeBulk(nil)
		} else {
			return err
		}
	} else {
		c.resp.writeBulk(num.FormatInt64ToSlice(s))
	}

	return nil
}

func zremCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	n, err := c.db.ZRem(args[0], args[1:]...)

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
}

func zincrbyCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	key := args[0]

	delta, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	v, err := c.db.ZIncrBy(key, delta, args[2])

	if err == nil {
		c.resp.writeBulk(num.FormatInt64ToSlice(v))
	}

	return err
}

func zparseScoreRange(minBuf []byte, maxBuf []byte) (min int64, max int64, err error) {
	if strings.ToLower(hack.String(minBuf)) == "-inf" {
		min = math.MinInt64
	} else {

		if len(minBuf) == 0 {
			err = ErrCmdParams
			return
		}

		var lopen bool = false
		if minBuf[0] == '(' {
			lopen = true
			minBuf = minBuf[1:]
		}

		min, err = ledis.StrInt64(minBuf, nil)
		if err != nil {
			err = ErrValue
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

	if strings.ToLower(hack.String(maxBuf)) == "+inf" {
		max = math.MaxInt64
	} else {
		var ropen = false

		if len(maxBuf) == 0 {
			err = ErrCmdParams
			return
		}
		if maxBuf[0] == '(' {
			ropen = true
			maxBuf = maxBuf[1:]
		}

		if maxBuf[0] == '(' {
			ropen = true
			maxBuf = maxBuf[1:]
		}

		max, err = ledis.StrInt64(maxBuf, nil)
		if err != nil {
			err = ErrValue
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
		return ErrValue
	}

	if min > max {
		c.resp.writeInteger(0)
		return nil
	}

	if n, err := c.db.ZCount(args[0], min, max); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
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
		c.resp.writeBulk(nil)
	} else {
		c.resp.writeInteger(n)
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
		c.resp.writeBulk(nil)
	} else {
		c.resp.writeInteger(n)
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
		return ErrValue
	}

	n, err := c.db.ZRemRangeByRank(key, start, stop)

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
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

	n, err := c.db.ZRemRangeByScore(key, min, max)

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
}

func zparseRange(c *client, a1 []byte, a2 []byte) (start int, stop int, err error) {
	if start, err = strconv.Atoi(hack.String(a1)); err != nil {
		return
	}

	if stop, err = strconv.Atoi(hack.String(a2)); err != nil {
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
		return ErrValue
	}

	args = args[3:]
	var withScores bool = false

	if len(args) > 0 {
		if len(args) != 1 {
			return ErrCmdParams
		}
		if strings.ToLower(hack.String(args[0])) == "withscores" {
			withScores = true
		} else {
			return ErrSyntax
		}
	}

	if datas, err := c.db.ZRangeGeneric(key, start, stop, reverse); err != nil {
		return err
	} else {
		c.resp.writeScorePairArray(datas, withScores)
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

	var minScore, maxScore []byte

	if !reverse {
		minScore, maxScore = args[1], args[2]
	} else {
		minScore, maxScore = args[2], args[1]
	}

	min, max, err := zparseScoreRange(minScore, maxScore)

	if err != nil {
		return err
	}

	args = args[3:]

	var withScores bool = false

	if len(args) > 0 {
		if strings.ToLower(hack.String(args[0])) == "withscores" {
			withScores = true
			args = args[1:]
		}
	}

	var offset int = 0
	var count int = -1

	if len(args) > 0 {
		if len(args) != 3 {
			return ErrCmdParams
		}

		if strings.ToLower(hack.String(args[0])) != "limit" {
			return ErrSyntax
		}

		if offset, err = strconv.Atoi(hack.String(args[1])); err != nil {
			return ErrValue
		}

		if count, err = strconv.Atoi(hack.String(args[2])); err != nil {
			return ErrValue
		}
	}

	if offset < 0 {
		//for ledis, if offset < 0, a empty will return
		//so here we directly return a empty array
		c.resp.writeArray([]interface{}{})
		return nil
	}

	if datas, err := c.db.ZRangeByScoreGeneric(key, min, max, offset, count, reverse); err != nil {
		return err
	} else {
		c.resp.writeScorePairArray(datas, withScores)
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

	n, err := c.db.ZClear(args[0])

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
}

func zmclearCommand(c *client) error {
	args := c.args
	if len(args) < 1 {
		return ErrCmdParams
	}

	n, err := c.db.ZMclear(args...)

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
}

func zexpireCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	duration, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	v, err := c.db.ZExpire(args[0], duration)

	if err == nil {
		c.resp.writeInteger(v)
	}

	return err
}

func zexpireAtCommand(c *client) error {
	args := c.args
	if len(args) != 2 {
		return ErrCmdParams
	}

	when, err := ledis.StrInt64(args[1], nil)
	if err != nil {
		return ErrValue
	}

	v, err := c.db.ZExpireAt(args[0], when)

	if err == nil {
		c.resp.writeInteger(v)
	}

	return err
}

func zttlCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	if v, err := c.db.ZTTL(args[0]); err != nil {
		return err
	} else {
		c.resp.writeInteger(v)
	}

	return nil
}

func zpersistCommand(c *client) error {
	args := c.args
	if len(args) != 1 {
		return ErrCmdParams
	}

	n, err := c.db.ZPersist(args[0])

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
}

func zparseZsetoptStore(args [][]byte) (destKey []byte, srcKeys [][]byte, weights []int64, aggregate byte, err error) {
	destKey = args[0]
	nKeys, err := strconv.Atoi(hack.String(args[1]))
	if err != nil {
		err = ErrValue
		return
	}
	args = args[2:]
	if len(args) < nKeys {
		err = ErrSyntax
		return
	}

	srcKeys = args[:nKeys]

	args = args[nKeys:]

	var weightsFlag = false
	var aggregateFlag = false

	for len(args) > 0 {
		if strings.ToLower(hack.String(args[0])) == "weights" {
			if weightsFlag {
				err = ErrSyntax
				return
			}

			args = args[1:]
			if len(args) < nKeys {
				err = ErrSyntax
				return
			}

			weights = make([]int64, nKeys)
			for i, arg := range args[:nKeys] {
				if weights[i], err = ledis.StrInt64(arg, nil); err != nil {
					err = ErrValue
					return
				}
			}
			args = args[nKeys:]

			weightsFlag = true

		} else if strings.ToLower(hack.String(args[0])) == "aggregate" {
			if aggregateFlag {
				err = ErrSyntax
				return
			}
			if len(args) < 2 {
				err = ErrSyntax
				return
			}

			if strings.ToLower(hack.String(args[1])) == "sum" {
				aggregate = ledis.AggregateSum
			} else if strings.ToLower(hack.String(args[1])) == "min" {
				aggregate = ledis.AggregateMin
			} else if strings.ToLower(hack.String(args[1])) == "max" {
				aggregate = ledis.AggregateMax
			} else {
				err = ErrSyntax
				return
			}
			args = args[2:]
			aggregateFlag = true
		} else {
			err = ErrSyntax
			return
		}
	}
	if !aggregateFlag {
		aggregate = ledis.AggregateSum
	}
	return
}

func zunionstoreCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	destKey, srcKeys, weights, aggregate, err := zparseZsetoptStore(args)
	if err != nil {
		return err
	}

	n, err := c.db.ZUnionStore(destKey, srcKeys, weights, aggregate)

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
}

func zinterstoreCommand(c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	destKey, srcKeys, weights, aggregate, err := zparseZsetoptStore(args)
	if err != nil {
		return err
	}

	n, err := c.db.ZInterStore(destKey, srcKeys, weights, aggregate)

	if err == nil {
		c.resp.writeInteger(n)
	}

	return err
}

func zxscanCommand(c *client) error {
	return xscanGeneric(c, c.db.ZScan)
}

func zxrevscanCommand(c *client) error {
	return xscanGeneric(c, c.db.ZRevScan)
}

func zparseMemberRange(minBuf []byte, maxBuf []byte) (min []byte, max []byte, rangeType uint8, err error) {
	rangeType = store.RangeClose
	if strings.ToLower(hack.String(minBuf)) == "-" {
		min = nil
	} else {
		if len(minBuf) == 0 {
			err = ErrCmdParams
			return
		}

		if minBuf[0] == '(' {
			rangeType |= store.RangeLOpen
			min = minBuf[1:]
		} else if minBuf[0] == '[' {
			min = minBuf[1:]
		} else {
			err = ErrCmdParams
			return
		}
	}

	if strings.ToLower(hack.String(maxBuf)) == "+" {
		max = nil
	} else {
		if len(maxBuf) == 0 {
			err = ErrCmdParams
			return
		}
		if maxBuf[0] == '(' {
			rangeType |= store.RangeROpen
			max = maxBuf[1:]
		} else if maxBuf[0] == '[' {
			max = maxBuf[1:]
		} else {
			err = ErrCmdParams
			return
		}
	}

	return
}

func zrangebylexCommand(c *client) error {
	args := c.args
	if len(args) != 3 && len(args) != 6 {
		return ErrCmdParams
	}

	min, max, rangeType, err := zparseMemberRange(args[1], args[2])
	if err != nil {
		return err
	}

	var offset int = 0
	var count int = -1

	if len(args) == 6 {
		if strings.ToLower(hack.String(args[3])) != "limit" {
			return ErrSyntax
		}

		if offset, err = strconv.Atoi(hack.String(args[4])); err != nil {
			return ErrValue
		}

		if count, err = strconv.Atoi(hack.String(args[5])); err != nil {
			return ErrValue
		}
	}

	key := args[0]
	if ay, err := c.db.ZRangeByLex(key, min, max, rangeType, offset, count); err != nil {
		return err
	} else {
		c.resp.writeSliceArray(ay)
	}

	return nil
}

func zremrangebylexCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	min, max, rangeType, err := zparseMemberRange(args[1], args[2])
	if err != nil {
		return err
	}

	key := args[0]
	if n, err := c.db.ZRemRangeByLex(key, min, max, rangeType); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
	}

	return nil
}

func zlexcountCommand(c *client) error {
	args := c.args
	if len(args) != 3 {
		return ErrCmdParams
	}

	min, max, rangeType, err := zparseMemberRange(args[1], args[2])
	if err != nil {
		return err
	}

	key := args[0]
	if n, err := c.db.ZLexCount(key, min, max, rangeType); err != nil {
		return err
	} else {
		c.resp.writeInteger(n)
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

	register("zunionstore", zunionstoreCommand)
	register("zinterstore", zinterstoreCommand)

	register("zrangebylex", zrangebylexCommand)
	register("zremrangebylex", zremrangebylexCommand)
	register("zlexcount", zlexcountCommand)

	//ledisdb special command

	register("zclear", zclearCommand)
	register("zmclear", zmclearCommand)
	register("zexpire", zexpireCommand)
	register("zexpireat", zexpireAtCommand)
	register("zttl", zttlCommand)
	register("zpersist", zpersistCommand)
	register("zxscan", zxscanCommand)
	register("zxrevscan", zxrevscanCommand)
}
