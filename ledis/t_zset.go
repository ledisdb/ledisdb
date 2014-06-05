package ledis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/siddontang/go-leveldb/leveldb"
)

const (
	MinScore     int64 = -1<<63 + 1
	MaxScore     int64 = 1<<63 - 1
	InvalidScore int64 = -1 << 63
)

type ScorePair struct {
	Score  int64
	Member []byte
}

var errZSizeKey = errors.New("invalid zsize key")
var errZSetKey = errors.New("invalid zset key")
var errZScoreKey = errors.New("invalid zscore key")
var errScoreOverflow = errors.New("zset score overflow")

const (
	zsetNScoreSep    byte = '<'
	zsetPScoreSep    byte = zsetNScoreSep + 1
	zsetStopScoreSep byte = zsetPScoreSep + 1

	zsetStartMemSep byte = ':'
	zsetStopMemSep  byte = zsetStartMemSep + 1
)

func checkZSetKMSize(key []byte, member []byte) error {
	if len(key) > MaxKeySize || len(key) == 0 {
		return errKeySize
	} else if len(member) > MaxZSetMemberSize || len(member) == 0 {
		return errZSetMemberSize
	}
	return nil
}

func (db *DB) zEncodeSizeKey(key []byte) []byte {
	buf := make([]byte, len(key)+2)
	buf[0] = db.index
	buf[1] = zSizeType

	copy(buf[2:], key)
	return buf
}

func (db *DB) zDecodeSizeKey(ek []byte) ([]byte, error) {
	if len(ek) < 2 || ek[0] != db.index || ek[1] != zSizeType {
		return nil, errZSizeKey
	}

	return ek[2:], nil
}

func (db *DB) zEncodeSetKey(key []byte, member []byte) []byte {
	buf := make([]byte, len(key)+len(member)+5)

	pos := 0
	buf[pos] = db.index
	pos++

	buf[pos] = zsetType
	pos++

	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2

	copy(buf[pos:], key)
	pos += len(key)

	buf[pos] = zsetStartMemSep
	pos++

	copy(buf[pos:], member)

	return buf
}

func (db *DB) zDecodeSetKey(ek []byte) ([]byte, []byte, error) {
	if len(ek) < 5 || ek[0] != db.index || ek[1] != zsetType {
		return nil, nil, errZSetKey
	}

	keyLen := int(binary.BigEndian.Uint16(ek[2:]))
	if keyLen+5 > len(ek) {
		return nil, nil, errZSetKey
	}

	key := ek[4 : 4+keyLen]

	if ek[4+keyLen] != zsetStartMemSep {
		return nil, nil, errZSetKey
	}

	member := ek[5+keyLen:]
	return key, member, nil
}

func (db *DB) zEncodeStartSetKey(key []byte) []byte {
	k := db.zEncodeSetKey(key, nil)
	return k
}

func (db *DB) zEncodeStopSetKey(key []byte) []byte {
	k := db.zEncodeSetKey(key, nil)
	k[len(k)-1] = zsetStartMemSep + 1
	return k
}

func (db *DB) zEncodeScoreKey(key []byte, member []byte, score int64) []byte {
	buf := make([]byte, len(key)+len(member)+14)

	pos := 0
	buf[pos] = db.index
	pos++

	buf[pos] = zScoreType
	pos++

	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2

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

func (db *DB) zEncodeStartScoreKey(key []byte, score int64) []byte {
	return db.zEncodeScoreKey(key, nil, score)
}

func (db *DB) zEncodeStopScoreKey(key []byte, score int64) []byte {
	k := db.zEncodeScoreKey(key, nil, score)
	k[len(k)-1] = zsetStopMemSep
	return k
}

func (db *DB) zDecodeScoreKey(ek []byte) (key []byte, member []byte, score int64, err error) {
	if len(ek) < 14 || ek[0] != db.index || ek[1] != zScoreType {
		err = errZScoreKey
		return
	}

	keyLen := int(binary.BigEndian.Uint16(ek[2:]))
	if keyLen+14 > len(ek) {
		err = errZScoreKey
		return
	}

	key = ek[4 : 4+keyLen]
	pos := 4 + keyLen

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
	ek := db.zEncodeSetKey(key, member)

	if v, err := db.db.Get(ek); err != nil {
		return 0, err
	} else if v != nil {
		exists = 1

		if s, err := Int64(v, err); err != nil {
			return 0, err
		} else {
			sk := db.zEncodeScoreKey(key, member, s)
			t.Delete(sk)
		}
	}

	t.Put(ek, PutInt64(score))

	sk := db.zEncodeScoreKey(key, member, score)
	t.Put(sk, []byte{})

	return exists, nil
}

func (db *DB) zDelItem(key []byte, member []byte, skipDelScore bool) (int64, error) {
	t := db.zsetTx

	ek := db.zEncodeSetKey(key, member)
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
				sk := db.zEncodeScoreKey(key, member, s)
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

		if err := checkZSetKMSize(key, member); err != nil {
			return 0, err
		}

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
	sk := db.zEncodeSizeKey(key)

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
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	sk := db.zEncodeSizeKey(key)
	return Int64(db.db.Get(sk))
}

func (db *DB) ZScore(key []byte, member []byte) (int64, error) {
	if err := checkZSetKMSize(key, member); err != nil {
		return InvalidScore, err
	}

	var score int64 = InvalidScore

	k := db.zEncodeSetKey(key, member)
	if v, err := db.db.Get(k); err != nil {
		return InvalidScore, err
	} else if v == nil {
		return InvalidScore, ErrScoreMiss
	} else {
		if score, err = Int64(v, nil); err != nil {
			return InvalidScore, err
		}
	}

	return score, nil
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
		if err := checkZSetKMSize(key, members[i]); err != nil {
			return 0, err
		}

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

func (db *DB) ZIncrBy(key []byte, delta int64, member []byte) (int64, error) {
	if err := checkZSetKMSize(key, member); err != nil {
		return InvalidScore, err
	}

	t := db.zsetTx
	t.Lock()
	defer t.Unlock()

	ek := db.zEncodeSetKey(key, member)

	var score int64 = delta

	v, err := db.db.Get(ek)
	if err != nil {
		return InvalidScore, err
	} else if v != nil {
		if s, err := Int64(v, err); err != nil {
			return InvalidScore, err
		} else {
			sk := db.zEncodeScoreKey(key, member, s)
			t.Delete(sk)

			score = s + delta
			if score >= MaxScore || score <= MinScore {
				return InvalidScore, errScoreOverflow
			}
		}
	} else {
		db.zIncrSize(key, 1)
	}

	t.Put(ek, PutInt64(score))

	sk := db.zEncodeScoreKey(key, member, score)
	t.Put(sk, []byte{})

	err = t.Commit()
	return score, err
}

func (db *DB) ZCount(key []byte, min int64, max int64) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}
	minKey := db.zEncodeStartScoreKey(key, min)
	maxKey := db.zEncodeStopScoreKey(key, max)

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
	if err := checkZSetKMSize(key, member); err != nil {
		return 0, err
	}

	k := db.zEncodeSetKey(key, member)

	if v, err := db.db.Get(k); err != nil {
		return 0, err
	} else if v == nil {
		return -1, nil
	} else {
		if s, err := Int64(v, err); err != nil {
			return 0, err
		} else {
			var it *leveldb.Iterator

			sk := db.zEncodeScoreKey(key, member, s)

			if !reverse {
				minKey := db.zEncodeStartScoreKey(key, MinScore)
				it = db.db.Iterator(minKey, sk, leveldb.RangeClose, 0, -1)
			} else {
				maxKey := db.zEncodeStopScoreKey(key, MaxScore)
				it = db.db.RevIterator(sk, maxKey, leveldb.RangeClose, 0, -1)
			}

			var lastKey []byte = nil
			var n int64 = 0

			for ; it.Valid(); it.Next() {
				n++

				lastKey = it.Key()
			}

			it.Close()

			if _, m, _, err := db.zDecodeScoreKey(lastKey); err == nil && bytes.Equal(m, member) {
				n--
				return n, nil
			}
		}
	}

	return -1, nil
}

func (db *DB) zIterator(key []byte, min int64, max int64, offset int, limit int, reverse bool) *leveldb.Iterator {
	minKey := db.zEncodeStartScoreKey(key, min)
	maxKey := db.zEncodeStopScoreKey(key, max)

	if !reverse {
		return db.db.Iterator(minKey, maxKey, leveldb.RangeClose, offset, limit)
	} else {
		return db.db.RevIterator(minKey, maxKey, leveldb.RangeClose, offset, limit)
	}
}

func (db *DB) zRemRange(key []byte, min int64, max int64, offset int, limit int) (int64, error) {
	if len(key) > MaxKeySize {
		return 0, errKeySize
	}

	t := db.zsetTx
	t.Lock()
	defer t.Unlock()

	it := db.zIterator(key, min, max, offset, limit, false)
	var num int64 = 0
	for ; it.Valid(); it.Next() {
		k := it.Key()
		_, m, _, err := db.zDecodeScoreKey(k)
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
	it.Close()

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
	if len(key) > MaxKeySize {
		return nil, errKeySize
	}

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
		_, m, s, err := db.zDecodeScoreKey(it.Key())
		//may be we will check key equal?
		if err != nil {
			continue
		}

		if withScores {
			v = append(v, m, s)
		} else {
			v = append(v, m)
		}
	}
	it.Close()

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

func (db *DB) ZFlush() (drop int64, err error) {
	t := db.zsetTx
	t.Lock()
	defer t.Unlock()

	minKey := make([]byte, 2)
	minKey[0] = db.index
	minKey[1] = zsetType

	maxKey := make([]byte, 2)
	maxKey[0] = db.index
	maxKey[1] = zScoreType + 1

	it := db.db.Iterator(minKey, maxKey, leveldb.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		drop++
		if drop%1000 == 0 {
			if err = t.Commit(); err != nil {
				return
			}
		}
	}
	it.Close()

	err = t.Commit()
	// to do : binlog
	return
}

func (db *DB) ZScan(key []byte, member []byte, count int, inclusive bool) ([]ScorePair, error) {
	var minKey []byte
	if member != nil {
		if err := checkZSetKMSize(key, member); err != nil {
			return nil, err
		}

		minKey = db.zEncodeSetKey(key, member)
	} else {
		minKey = db.zEncodeStartSetKey(key)
	}

	maxKey := db.zEncodeStopSetKey(key)

	if count <= 0 {
		count = defaultScanCount
	}

	v := make([]ScorePair, 0, 2*count)

	rangeType := leveldb.RangeROpen
	if !inclusive {
		rangeType = leveldb.RangeOpen
	}

	it := db.db.Iterator(minKey, maxKey, rangeType, 0, count)
	for ; it.Valid(); it.Next() {
		if _, m, err := db.zDecodeSetKey(it.Key()); err != nil {
			continue
		} else {
			score, _ := Int64(it.Value(), nil)
			v = append(v, ScorePair{Member: m, Score: score})
		}
	}
	it.Close()

	return v, nil
}
