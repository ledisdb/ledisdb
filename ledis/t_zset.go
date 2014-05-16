package ledis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/siddontang/go-leveldb/leveldb"
)

const (
	MinScore int64 = -1<<63 + 1
	MaxScore int64 = 1<<63 - 1
)

type ScorePair struct {
	Score  int64
	Member []byte
}

var errZSizeKey = errors.New("invalid zsize key")
var errZSetKey = errors.New("invalid zset key")
var errZScoreKey = errors.New("invalid zscore key")

const (
	zsetNScoreSep    byte = '<'
	zsetPScoreSep    byte = zsetNScoreSep + 1
	zsetStopScoreSep byte = zsetPScoreSep + 1

	zsetStartMemSep byte = ':'
	zsetStopMemSep  byte = zsetStartMemSep + 1
)

func encode_zsize_key(key []byte) []byte {
	buf := make([]byte, len(key)+1)
	buf[0] = ZSIZE_TYPE

	copy(buf[1:], key)
	return buf
}

func decode_zsize_key(ek []byte) ([]byte, error) {
	if len(ek) == 0 || ek[0] != ZSIZE_TYPE {
		return nil, errZSizeKey
	}

	return ek[1:], nil
}

func encode_zset_key(key []byte, member []byte) []byte {
	buf := make([]byte, len(key)+len(member)+5)

	pos := 0
	buf[pos] = ZSET_TYPE
	pos++

	binary.BigEndian.PutUint32(buf[pos:], uint32(len(key)))
	pos += 4

	copy(buf[pos:], key)
	pos += len(key)

	copy(buf[pos:], member)

	return buf
}

func decode_zset_key(ek []byte) ([]byte, []byte, error) {
	if len(ek) < 5 || ek[0] != ZSET_TYPE {
		return nil, nil, errZSetKey
	}

	keyLen := int(binary.BigEndian.Uint32(ek[1:]))
	if keyLen+5 > len(ek) {
		return nil, nil, errZSetKey
	}

	key := ek[5 : 5+keyLen]
	member := ek[5+keyLen:]
	return key, member, nil
}

func encode_zscore_key(key []byte, member []byte, score int64) []byte {
	buf := make([]byte, len(key)+len(member)+15)

	pos := 0
	buf[pos] = ZSCORE_TYPE
	pos++

	binary.BigEndian.PutUint32(buf[pos:], uint32(len(key)))
	pos += 4

	copy(buf[pos:], key)
	pos += len(key)

	if score < 0 {
		buf[pos] = zsetNScoreSep
	} else {
		buf[pos] = zsetPScoreSep
	}

	pos++
	binary.BigEndian.PutUint64(buf[pos:], uint64(score))
	pos += 8

	buf[pos] = zsetStartMemSep
	pos++

	copy(buf[pos:], member)
	return buf
}

func encode_start_zscore_key(key []byte, score int64) []byte {
	k := encode_zscore_key(key, nil, score)
	return k
}

func encode_stop_zscore_key(key []byte, score int64) []byte {
	k := encode_zscore_key(key, nil, score)
	k[len(k)-1] = zsetStopMemSep
	return k
}

func decode_zscore_key(ek []byte) (key []byte, member []byte, score int64, err error) {
	if len(ek) < 15 || ek[0] != ZSCORE_TYPE {
		err = errZScoreKey
		return
	}

	keyLen := int(binary.BigEndian.Uint32(ek[1:]))
	if keyLen+14 > len(ek) {
		err = errZScoreKey
		return
	}

	key = ek[5 : 5+keyLen]
	pos := 5 + keyLen

	if (ek[pos] != zsetNScoreSep) && (ek[pos] != zsetPScoreSep) {
		err = errZScoreKey
		return
	}
	pos++

	score = int64(binary.BigEndian.Uint64(ek[pos:]))
	pos += 8

	if ek[pos] != zsetStartMemSep {
		err = errZScoreKey
		return
	}

	pos++

	member = ek[pos:]
	return
}

func (db *DB) zSetItem(key []byte, score int64, member []byte) (int64, error) {
	if score <= MinScore || score >= MaxScore {
		return 0, errScoreOverflow
	}

	t := db.zsetTx

	var exists int64 = 0
	ek := encode_zset_key(key, member)
	if v, err := db.db.Get(ek); err != nil {
		return 0, err
	} else if v != nil {
		exists = 1

		if s, err := Int64(v, err); err != nil {
			return 0, err
		} else {
			sk := encode_zscore_key(key, member, s)
			t.Delete(sk)
		}
	}

	t.Put(ek, PutInt64(score))

	sk := encode_zscore_key(key, member, score)
	t.Put(sk, []byte{})

	return exists, nil
}

func (db *DB) zDelItem(key []byte, member []byte, skipDelScore bool) (int64, error) {
	t := db.zsetTx

	ek := encode_zset_key(key, member)
	if v, err := db.db.Get(ek); err != nil {
		return 0, err
	} else if v == nil {
		//not exists
		return 0, nil
	} else {
		//exists
		if !skipDelScore {
			//we must del score
			if s, err := Int64(v, err); err != nil {
				return 0, err
			} else {
				sk := encode_zscore_key(key, member, s)
				t.Delete(sk)
			}
		}
	}

	t.Delete(ek)
	return 1, nil
}

func (db *DB) ZAdd(key []byte, args ...ScorePair) (int64, error) {
	if len(args) == 0 {
		return 0, nil
	}

	t := db.zsetTx
	t.Lock()
	defer t.Unlock()

	var num int64 = 0
	for i := 0; i < len(args); i++ {
		score := args[i].Score
		member := args[i].Member

		if n, err := db.zSetItem(key, score, member); err != nil {
			return 0, err
		} else if n == 0 {
			//add new
			num++
		}
	}

	if _, err := db.zIncrSize(key, num); err != nil {
		return 0, err
	}

	//todo add binlog
	err := t.Commit()
	return num, err
}

func (db *DB) zIncrSize(key []byte, delta int64) (int64, error) {
	t := db.zsetTx
	sk := encode_zsize_key(key)
	size, err := Int64(db.db.Get(sk))
	if err != nil {
		return 0, err
	} else {
		size += delta
		if size <= 0 {
			size = 0
			t.Delete(sk)
		} else {
			t.Put(sk, PutInt64(size))
		}
	}

	return size, nil
}

func (db *DB) ZCard(key []byte) (int64, error) {
	sk := encode_zsize_key(key)
	size, err := Int64(db.db.Get(sk))
	return size, err
}

func (db *DB) ZScore(key []byte, member []byte) ([]byte, error) {
	k := encode_zset_key(key, member)
	score, err := Int64(db.db.Get(k))
	if err != nil {
		return nil, err
	}

	return StrPutInt64(score), nil
}

func (db *DB) ZRem(key []byte, members ...[]byte) (int64, error) {
	if len(members) == 0 {
		return 0, nil
	}

	t := db.zsetTx
	t.Lock()
	defer t.Unlock()

	var num int64 = 0
	for i := 0; i < len(members); i++ {
		if n, err := db.zDelItem(key, members[i], false); err != nil {
			return 0, err
		} else if n == 1 {
			num++
		}
	}

	if _, err := db.zIncrSize(key, -num); err != nil {
		return 0, err
	}

	err := t.Commit()
	return num, err
}

func (db *DB) ZIncrBy(key []byte, delta int64, member []byte) ([]byte, error) {
	t := db.zsetTx
	t.Lock()
	defer t.Unlock()

	ek := encode_zset_key(key, member)
	var score int64 = delta
	v, err := db.db.Get(ek)
	if err != nil {
		return nil, err
	} else if v != nil {
		if s, err := Int64(v, err); err != nil {
			return nil, err
		} else {
			sk := encode_zscore_key(key, member, s)
			t.Delete(sk)

			score = s + delta

			if score >= MaxScore || score <= MinScore {
				return nil, errScoreOverflow
			}
		}
	} else {
		db.zIncrSize(key, 1)
	}

	t.Put(ek, PutInt64(score))

	t.Put(encode_zscore_key(key, member, score), []byte{})

	err = t.Commit()
	return StrPutInt64(score), err
}

func (db *DB) ZCount(key []byte, min int64, max int64) (int64, error) {
	minKey := encode_start_zscore_key(key, min)
	maxKey := encode_stop_zscore_key(key, max)

	rangeType := leveldb.RangeROpen

	it := db.db.Iterator(minKey, maxKey, rangeType, 0, -1)
	var n int64 = 0
	for ; it.Valid(); it.Next() {
		n++
	}
	it.Close()

	return n, nil
}

func (db *DB) zrank(key []byte, member []byte, reverse bool) (int64, error) {
	k := encode_zset_key(key, member)

	if v, err := db.db.Get(k); err != nil {
		return 0, err
	} else if v == nil {
		return -1, nil
	} else {
		if s, err := Int64(v, err); err != nil {
			return 0, err
		} else {
			var it *leveldb.Iterator

			sk := encode_zscore_key(key, member, s)

			if !reverse {
				minKey := encode_start_zscore_key(key, MinScore)
				it = db.db.Iterator(minKey, sk, leveldb.RangeClose, 0, -1)
			} else {
				maxKey := encode_stop_zscore_key(key, MaxScore)
				it = db.db.RevIterator(sk, maxKey, leveldb.RangeClose, 0, -1)
			}

			var lastKey []byte = nil
			var n int64 = 0

			for ; it.Valid(); it.Next() {
				n++

				lastKey = it.Key()
			}

			it.Close()

			if _, m, _, err := decode_zscore_key(lastKey); err == nil && bytes.Equal(m, member) {
				n--
				return n, nil
			}
		}
	}

	return -1, nil
}

func (db *DB) zIterator(key []byte, min int64, max int64, offset int, limit int, reverse bool) *leveldb.Iterator {
	minKey := encode_start_zscore_key(key, min)
	maxKey := encode_stop_zscore_key(key, max)

	if !reverse {
		return db.db.Iterator(minKey, maxKey, leveldb.RangeClose, offset, limit)
	} else {
		return db.db.RevIterator(minKey, maxKey, leveldb.RangeClose, offset, limit)
	}
}

func (db *DB) zRemRange(key []byte, min int64, max int64, offset int, limit int) (int64, error) {
	t := db.zsetTx
	t.Lock()
	defer t.Unlock()

	it := db.zIterator(key, min, max, offset, limit, false)
	var num int64 = 0
	for ; it.Valid(); it.Next() {
		k := it.Key()
		_, m, _, err := decode_zscore_key(k)
		if err != nil {
			continue
		}

		if n, err := db.zDelItem(key, m, true); err != nil {
			return 0, err
		} else if n == 1 {
			num++
		}

		t.Delete(k)
	}

	if _, err := db.zIncrSize(key, -num); err != nil {
		return 0, err
	}

	//todo add binlog

	err := t.Commit()
	return num, err
}

func (db *DB) zReverse(s []interface{}, withScores bool) []interface{} {
	if withScores {
		for i, j := 0, len(s)-2; i < j; i, j = i+2, j-2 {
			s[i], s[j] = s[j], s[i]
			s[i+1], s[j+1] = s[j+1], s[i+1]
		}
	} else {
		for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
			s[i], s[j] = s[j], s[i]
		}
	}

	return s
}

func (db *DB) zRange(key []byte, min int64, max int64, withScores bool, offset int, limit int, reverse bool) ([]interface{}, error) {
	if offset < 0 {
		return []interface{}{}, nil
	}

	nv := 64
	if limit > 0 {
		nv = limit
	}
	if withScores {
		nv = 2 * nv
	}
	v := make([]interface{}, 0, nv)

	var it *leveldb.Iterator

	//if reverse and offset is 0, limit < 0, we may use forward iterator then reverse
	//because leveldb iterator prev is slower than next
	if !reverse || (offset == 0 && limit < 0) {
		it = db.zIterator(key, min, max, offset, limit, false)
	} else {
		it = db.zIterator(key, min, max, offset, limit, true)
	}

	for ; it.Valid(); it.Next() {
		_, m, s, err := decode_zscore_key(it.Key())
		//may be we will check key equal?
		if err != nil {
			continue
		}

		v = append(v, m)

		if withScores {
			v = append(v, StrPutInt64(s))
		}
	}

	if reverse && (offset == 0 && limit < 0) {
		v = db.zReverse(v, withScores)
	}

	return v, nil
}

func (db *DB) zParseLimit(key []byte, start int, stop int) (offset int, limit int, err error) {
	if start < 0 || stop < 0 {
		//refer redis implementation
		var size int64
		size, err = db.ZCard(key)
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

func (db *DB) ZClear(key []byte) (int64, error) {
	return db.zRemRange(key, MinScore, MaxScore, 0, -1)
}

func (db *DB) ZRange(key []byte, start int, stop int, withScores bool) ([]interface{}, error) {
	return db.ZRangeGeneric(key, start, stop, withScores, false)
}

//min and max must be inclusive
//if no limit, set offset = 0 and count = -1
func (db *DB) ZRangeByScore(key []byte, min int64, max int64,
	withScores bool, offset int, count int) ([]interface{}, error) {
	return db.ZRangeByScoreGeneric(key, min, max, withScores, offset, count, false)
}

func (db *DB) ZRank(key []byte, member []byte) (int64, error) {
	return db.zrank(key, member, false)
}

func (db *DB) ZRemRangeByRank(key []byte, start int, stop int) (int64, error) {
	offset, limit, err := db.zParseLimit(key, start, stop)
	if err != nil {
		return 0, err
	}
	return db.zRemRange(key, MinScore, MaxScore, offset, limit)
}

//min and max must be inclusive
func (db *DB) ZRemRangeByScore(key []byte, min int64, max int64) (int64, error) {
	return db.zRemRange(key, min, max, 0, -1)
}

func (db *DB) ZRevRange(key []byte, start int, stop int, withScores bool) ([]interface{}, error) {
	return db.ZRangeGeneric(key, start, stop, withScores, true)
}

func (db *DB) ZRevRank(key []byte, member []byte) (int64, error) {
	return db.zrank(key, member, true)
}

//min and max must be inclusive
//if no limit, set offset = 0 and count = -1
func (db *DB) ZRevRangeByScore(key []byte, min int64, max int64,
	withScores bool, offset int, count int) ([]interface{}, error) {
	return db.ZRangeByScoreGeneric(key, min, max, withScores, offset, count, true)
}

func (db *DB) ZRangeGeneric(key []byte, start int, stop int,
	withScores bool, reverse bool) ([]interface{}, error) {
	offset, limit, err := db.zParseLimit(key, start, stop)
	if err != nil {
		return nil, err
	}

	return db.zRange(key, MinScore, MaxScore, withScores, offset, limit, reverse)
}

//min and max must be inclusive
//if no limit, set offset = 0 and count = -1
func (db *DB) ZRangeByScoreGeneric(key []byte, min int64, max int64,
	withScores bool, offset int, count int, reverse bool) ([]interface{}, error) {

	return db.zRange(key, min, max, withScores, offset, count, reverse)
}
