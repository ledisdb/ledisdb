package http

import (
	"errors"
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"math"
	"strconv"
	"strings"
)

var errScoreOverflow = errors.New("zset score overflow")

func zaddCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zadd")
	}

	if len(args[1:])%2 != 0 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zadd")
	}

	key := []byte(args[0])
	args = args[1:]

	params := make([]ledis.ScorePair, len(args)/2)
	for i := 0; i < len(params); i++ {
		score, err := strconv.ParseInt(args[2*i], 10, 64)
		if err != nil {
			return nil, ErrValue
		}

		params[i].Score = score
		params[i].Member = []byte(args[2*i+1])
	}

	if n, err := db.ZAdd(key, params...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func zcardCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zcard")
	}

	key := []byte(args[0])
	if n, err := db.ZCard(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func zscoreCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zscore")
	}

	key := []byte(args[0])
	member := []byte(args[1])

	if s, err := db.ZScore(key, member); err != nil {
		if err == ledis.ErrScoreMiss {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return strconv.FormatInt(s, 10), nil
	}
}

func zremCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zrem")
	}

	key := []byte(args[0])
	members := make([][]byte, len(args[1:]))
	for i, arg := range args[1:] {
		members[i] = []byte(arg)
	}
	if n, err := db.ZRem(key, members...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func zincrbyCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zincrby")
	}

	key := []byte(args[0])

	delta, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	member := []byte(args[2])
	if v, err := db.ZIncrBy(key, delta, member); err != nil {
		return nil, err
	} else {
		return strconv.FormatInt(v, 10), nil
	}
}

func zparseScoreRange(minBuf string, maxBuf string) (min int64, max int64, err error) {
	if strings.ToLower(minBuf) == "-inf" {
		min = math.MinInt64
	} else {
		var lopen bool = false

		if len(minBuf) == 0 {
			err = ErrValue
			return
		}

		if minBuf[0] == '(' {
			lopen = true
			minBuf = minBuf[1:]
		}

		min, err = strconv.ParseInt(minBuf, 10, 64)
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

	if strings.ToLower(maxBuf) == "+inf" {
		max = math.MaxInt64
	} else {
		var ropen = false

		if len(maxBuf) == 0 {
			err = ErrValue
			return
		}

		if maxBuf[0] == '(' {
			ropen = true
			maxBuf = maxBuf[1:]
		}

		max, err = strconv.ParseInt(maxBuf, 10, 64)
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

func zcountCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zcount")
	}

	min, max, err := zparseScoreRange(args[1], args[2])
	if err != nil {
		return nil, err
	}

	if min > max {
		return 0, nil
	}

	key := []byte(args[0])
	if n, err := db.ZCount(key, min, max); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func zrankCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zrank")
	}
	key := []byte(args[0])
	member := []byte(args[1])

	if n, err := db.ZRank(key, member); err != nil {
		return nil, err
	} else if n == -1 {
		return nil, nil
	} else {
		return n, nil
	}
}

func zrevrankCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zrevrank")
	}

	key := []byte(args[0])
	member := []byte(args[1])
	if n, err := db.ZRevRank(key, member); err != nil {
		return nil, err
	} else if n == -1 {
		return nil, nil
	} else {
		return n, nil
	}
}

func zremrangebyrankCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zremrangebyrank")
	}

	key := []byte(args[0])

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, ErrValue
	}
	stop, err := strconv.Atoi(args[2])

	if err != nil {
		return nil, ErrValue
	}

	if n, err := db.ZRemRangeByRank(key, start, stop); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func zremrangebyscoreCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zremrangebyscore")
	}

	key := []byte(args[0])
	min, max, err := zparseScoreRange(args[1], args[2])
	if err != nil {
		return nil, err
	}

	if n, err := db.ZRemRangeByScore(key, min, max); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func zrangeGeneric(db *ledis.DB, reverse bool, args ...string) (interface{}, error) {

	key := []byte(args[0])

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, ErrValue
	}

	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, ErrValue
	}

	args = args[3:]
	var withScores bool = false

	if len(args) > 0 {
		if len(args) != 1 {
			return nil, ErrSyntax
		}
		if strings.ToLower(args[0]) == "withscores" {
			withScores = true
		} else {
			return nil, ErrSyntax
		}
	}

	if datas, err := db.ZRangeGeneric(key, start, stop, reverse); err != nil {
		return nil, err
	} else {
		return makeScorePairArray(datas, withScores), nil
	}
}

func makeScorePairArray(datas []ledis.ScorePair, withScores bool) []string {
	var arr []string
	if withScores {
		arr = make([]string, 2*len(datas))
		for i, data := range datas {
			arr[2*i] = ledis.String(data.Member)
			arr[2*i+1] = strconv.FormatInt(data.Score, 10)
		}
	} else {
		arr = make([]string, len(datas))
		for i, data := range datas {
			arr[i] = ledis.String(data.Member)
		}
	}
	return arr
}

func zrangeCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zrange")
	}
	return zrangeGeneric(db, false, args...)
}

func zrevrangeCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zrevrange")
	}
	return zrangeGeneric(db, true, args...)
}

func zrangebyscoreGeneric(db *ledis.DB, reverse bool, args ...string) (interface{}, error) {
	key := []byte(args[0])

	var minScore, maxScore string

	if !reverse {
		minScore, maxScore = args[1], args[2]
	} else {
		minScore, maxScore = args[2], args[1]
	}

	min, max, err := zparseScoreRange(minScore, maxScore)

	if err != nil {
		return nil, err
	}

	args = args[3:]

	var withScores bool = false

	if len(args) > 0 && strings.ToLower(args[0]) == "withscores" {
		withScores = true
		args = args[1:]
	}

	var offset int = 0
	var count int = -1

	if len(args) > 0 {
		if len(args) != 3 {
			return nil, ErrSyntax
		}

		if strings.ToLower(args[0]) != "limit" {
			return nil, ErrSyntax
		}

		if offset, err = strconv.Atoi(args[1]); err != nil {
			return nil, ErrValue
		}

		if count, err = strconv.Atoi(args[2]); err != nil {
			return nil, ErrValue
		}
	}

	if offset < 0 {
		return []interface{}{}, nil
	}

	if datas, err := db.ZRangeByScoreGeneric(key, min, max, offset, count, reverse); err != nil {
		return nil, err
	} else {
		return makeScorePairArray(datas, withScores), nil
	}
}

func zrangebyscoreCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zrangebyscore")
	}
	return zrangebyscoreGeneric(db, false, args...)
}

func zrevrangebyscoreCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zrevrangebyscore")
	}
	return zrangebyscoreGeneric(db, true, args...)
}

func zclearCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zclear")
	}

	key := []byte(args[0])
	if n, err := db.ZClear(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func zmclearCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zmclear")
	}

	keys := make([][]byte, len(args))
	for i, arg := range args {
		keys[i] = []byte(arg)
	}
	if n, err := db.ZMclear(keys...); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}

func zexpireCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zexpire")
	}

	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	key := []byte(args[0])
	if v, err := db.ZExpire(key, duration); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func zexpireAtCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zexpireat")
	}

	when, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrValue
	}

	key := []byte(args[0])
	if v, err := db.ZExpireAt(key, when); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func zttlCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zttl")
	}

	key := []byte(args[0])
	if v, err := db.ZTTL(key); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func zpersistCommand(db *ledis.DB, args ...string) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(ERR_ARGUMENT_FORMAT, "zpersist")
	}

	key := []byte(args[0])
	if n, err := db.ZPersist(key); err != nil {
		return nil, err
	} else {
		return n, nil
	}
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
