package ledis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/siddontang/golib/hack"
	"github.com/siddontang/golib/leveldb"
	"strconv"
)

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

func (a *App) zset_setItem(key []byte, score int64, member []byte) (int64, error) {
	if score <= MinScore || score >= MaxScore {
		return 0, errScoreOverflow
	}

	t := a.zsetTx

	var exists int64 = 0
	ek := encode_zset_key(key, member)
	if v, err := a.db.Get(ek); err != nil {
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

func (a *App) zset_delItem(key []byte, member []byte, skipDelScore bool) (int64, error) {
	t := a.zsetTx

	ek := encode_zset_key(key, member)
	if v, err := a.db.Get(ek); err != nil {
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

func (a *App) zset_add(key []byte, args []interface{}) (int64, error) {
	t := a.zsetTx
	t.Lock()
	defer t.Unlock()

	var num int64 = 0
	for i := 0; i < len(args); i += 2 {
		score := args[i].(int64)
		member := args[i+1].([]byte)

		if n, err := a.zset_setItem(key, score, member); err != nil {
			return 0, err
		} else if n == 0 {
			//add new
			num++
		}
	}

	if _, err := a.zset_incrSize(key, num); err != nil {
		return 0, err
	}

	//todo add binlog
	err := t.Commit()
	return num, err
}

func (a *App) zset_incrSize(key []byte, delta int64) (int64, error) {
	t := a.zsetTx
	sk := encode_zsize_key(key)
	size, err := Int64(a.db.Get(sk))
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

func (a *App) zset_card(key []byte) (int64, error) {
	sk := encode_zsize_key(key)
	size, err := Int64(a.db.Get(sk))
	return size, err
}

func (a *App) zset_score(key []byte, member []byte) ([]byte, error) {
	k := encode_zset_key(key, member)
	score, err := Int64(a.db.Get(k))
	if err != nil {
		return nil, err
	}

	return hack.Slice(strconv.FormatInt(score, 10)), nil
}

func (a *App) zset_rem(key []byte, args [][]byte) (int64, error) {
	t := a.zsetTx
	t.Lock()
	defer t.Unlock()

	var num int64 = 0
	for i := 0; i < len(args); i++ {
		if n, err := a.zset_delItem(key, args[i], false); err != nil {
			return 0, err
		} else if n == 1 {
			num++
		}
	}

	if _, err := a.zset_incrSize(key, -num); err != nil {
		return 0, err
	}

	err := t.Commit()
	return num, err
}

func (a *App) zset_incrby(key []byte, delta int64, member []byte) ([]byte, error) {
	t := a.zsetTx
	t.Lock()
	defer t.Unlock()

	ek := encode_zset_key(key, member)
	var score int64 = delta
	v, err := a.db.Get(ek)
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
		a.zset_incrSize(key, 1)
	}

	t.Put(ek, PutInt64(score))

	t.Put(encode_zscore_key(key, member, score), []byte{})

	err = t.Commit()
	return hack.Slice(strconv.FormatInt(score, 10)), err
}

func (a *App) zset_count(key []byte, min int64, max int64) (int64, error) {
	minKey := encode_start_zscore_key(key, min)
	maxKey := encode_stop_zscore_key(key, max)

	rangeType := leveldb.RangeROpen

	it := a.db.Iterator(minKey, maxKey, rangeType, 0, -1)
	var n int64 = 0
	for ; it.Valid(); it.Next() {
		n++
	}
	it.Close()

	return n, nil
}

func (a *App) zset_rank(key []byte, member []byte, reverse bool) (int64, error) {
	k := encode_zset_key(key, member)

	if v, err := a.db.Get(k); err != nil {
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
				it = a.db.Iterator(minKey, sk, leveldb.RangeClose, 0, -1)
			} else {
				maxKey := encode_stop_zscore_key(key, MaxScore)
				it = a.db.RevIterator(sk, maxKey, leveldb.RangeClose, 0, -1)
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

func (a *App) zset_iterator(key []byte, min int64, max int64, offset int, limit int, reverse bool) *leveldb.Iterator {
	minKey := encode_start_zscore_key(key, min)
	maxKey := encode_stop_zscore_key(key, max)

	if !reverse {
		return a.db.Iterator(minKey, maxKey, leveldb.RangeClose, offset, limit)
	} else {
		return a.db.RevIterator(minKey, maxKey, leveldb.RangeClose, offset, limit)
	}
}

func (a *App) zset_remRange(key []byte, min int64, max int64, offset int, limit int) (int64, error) {
	t := a.zsetTx
	t.Lock()
	defer t.Unlock()

	it := a.zset_iterator(key, min, max, offset, limit, false)
	var num int64 = 0
	for ; it.Valid(); it.Next() {
		k := it.Key()
		_, m, _, err := decode_zscore_key(k)
		if err != nil {
			continue
		}

		if n, err := a.zset_delItem(key, m, true); err != nil {
			return 0, err
		} else if n == 1 {
			num++
		}

		t.Delete(k)
	}

	if _, err := a.zset_incrSize(key, -num); err != nil {
		return 0, err
	}

	//todo add binlog

	err := t.Commit()
	return num, err
}

func (a *App) zset_range(key []byte, min int64, max int64, withScores bool, offset int, limit int, reverse bool) ([]interface{}, error) {
	v := make([]interface{}, 0, 16)
	it := a.zset_iterator(key, min, max, offset, limit, reverse)
	for ; it.Valid(); it.Next() {
		_, m, s, err := decode_zscore_key(it.Key())
		//may be we will check key equal?
		if err != nil {
			continue
		}

		v = append(v, m)

		if withScores {
			v = append(v, hack.Slice(strconv.FormatInt(s, 10)))
		}
	}

	return v, nil
}
