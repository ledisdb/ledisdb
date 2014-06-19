package ledis

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/ledisdb/leveldb"
	"time"
)

type FVPair struct {
	Field []byte
	Value []byte
}

var errHashKey = errors.New("invalid hash key")
var errHSizeKey = errors.New("invalid hsize key")

const (
	hashStartSep byte = ':'
	hashStopSep  byte = hashStartSep + 1
)

func checkHashKFSize(key []byte, field []byte) error {
	if len(key) > MaxKeySize || len(key) == 0 {
		return errKeySize
	} else if len(field) > MaxHashFieldSize || len(field) == 0 {
		return errHashFieldSize
	}
	return nil
}

func (db *DB) hEncodeSizeKey(key []byte) []byte {
	buf := make([]byte, len(key)+2)

	buf[0] = db.index
	buf[1] = hSizeType

	copy(buf[2:], key)
	return buf
}

func (db *DB) hDecodeSizeKey(ek []byte) ([]byte, error) {
	if len(ek) < 2 || ek[0] != db.index || ek[1] != hSizeType {
		return nil, errHSizeKey
	}

	return ek[2:], nil
}

func (db *DB) hEncodeHashKey(key []byte, field []byte) []byte {
	buf := make([]byte, len(key)+len(field)+1+1+2+1)

	pos := 0
	buf[pos] = db.index
	pos++
	buf[pos] = hashType
	pos++

	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2

	copy(buf[pos:], key)
	pos += len(key)

	buf[pos] = hashStartSep
	pos++
	copy(buf[pos:], field)

	return buf
}

func (db *DB) hDecodeHashKey(ek []byte) ([]byte, []byte, error) {
	if len(ek) < 5 || ek[0] != db.index || ek[1] != hashType {
		return nil, nil, errHashKey
	}

	pos := 2
	keyLen := int(binary.BigEndian.Uint16(ek[pos:]))
	pos += 2

	if keyLen+5 > len(ek) {
		return nil, nil, errHashKey
	}

	key := ek[pos : pos+keyLen]
	pos += keyLen

	if ek[pos] != hashStartSep {
		return nil, nil, errHashKey
	}

	pos++
	field := ek[pos:]
	return key, field, nil
}

func (db *DB) hEncodeStartKey(key []byte) []byte {
	return db.hEncodeHashKey(key, nil)
}

func (db *DB) hEncodeStopKey(key []byte) []byte {
	k := db.hEncodeHashKey(key, nil)

	k[len(k)-1] = hashStopSep

	return k
}

func (db *DB) hSetItem(key []byte, field []byte, value []byte) (int64, error) {
	t := db.hashTx

	ek := db.hEncodeHashKey(key, field)

	var n int64 = 1
	if v, _ := db.db.Get(ek); v != nil {
		n = 0
	} else {
		if _, err := db.hIncrSize(key, 1); err != nil {
			return 0, err
		}
	}

	t.Put(ek, value)
	return n, nil
}

//	ps : here just focus on deleting the hash data,
//		 any other likes expire is ignore.
func (db *DB) hDelete(t *tx, key []byte) int64 {
	sk := db.hEncodeSizeKey(key)
	start := db.hEncodeStartKey(key)
	stop := db.hEncodeStopKey(key)

	var num int64 = 0
	it := db.db.RangeLimitIterator(start, stop, leveldb.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		num++
	}
	it.Close()

	t.Delete(sk)
	return num
}

func (db *DB) hExpireAt(key []byte, when int64) (int64, error) {
	t := db.hashTx
	t.Lock()
	defer t.Unlock()

	if hlen, err := db.HLen(key); err != nil || hlen == 0 {
		return 0, err
	} else {
		db.expireAt(t, hExpType, key, when)
		if err := t.Commit(); err != nil {
			return 0, err
		}
	}
	return 1, nil
}

func (db *DB) HLen(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	return Int64(db.db.Get(db.hEncodeSizeKey(key)))
}

func (db *DB) HSet(key []byte, field []byte, value []byte) (int64, error) {
	if err := checkHashKFSize(key, field); err != nil {
		return 0, err
	} else if err := checkValueSize(value); err != nil {
		return 0, err
	}

	t := db.hashTx
	t.Lock()
	defer t.Unlock()

	n, err := db.hSetItem(key, field, value)
	if err != nil {
		return 0, err
	}

	//todo add binlog

	err = t.Commit()
	return n, err
}

func (db *DB) HGet(key []byte, field []byte) ([]byte, error) {
	if err := checkHashKFSize(key, field); err != nil {
		return nil, err
	}

	return db.db.Get(db.hEncodeHashKey(key, field))
}

func (db *DB) HMset(key []byte, args ...FVPair) error {
	t := db.hashTx
	t.Lock()
	defer t.Unlock()

	var err error
	var ek []byte
	var num int64 = 0
	for i := 0; i < len(args); i++ {
		if err := checkHashKFSize(key, args[i].Field); err != nil {
			return err
		} else if err := checkValueSize(args[i].Value); err != nil {
			return err
		}

		ek = db.hEncodeHashKey(key, args[i].Field)

		if v, err := db.db.Get(ek); err != nil {
			return err
		} else if v == nil {
			num++
		}

		t.Put(ek, args[i].Value)
	}

	if _, err = db.hIncrSize(key, num); err != nil {
		return err
	}

	//todo add binglog
	err = t.Commit()
	return err
}

func (db *DB) HMget(key []byte, args ...[]byte) ([]interface{}, error) {
	var ek []byte

	it := db.db.NewIterator()
	defer it.Close()

	r := make([]interface{}, len(args))
	for i := 0; i < len(args); i++ {
		if err := checkHashKFSize(key, args[i]); err != nil {
			return nil, err
		}

		ek = db.hEncodeHashKey(key, args[i])

		r[i] = it.Find(ek)
	}

	return r, nil
}

func (db *DB) HDel(key []byte, args [][]byte) (int64, error) {
	t := db.hashTx

	var ek []byte
	var v []byte
	var err error

	t.Lock()
	defer t.Unlock()

	var num int64 = 0
	for i := 0; i < len(args); i++ {
		if err := checkHashKFSize(key, args[i]); err != nil {
			return 0, err
		}

		ek = db.hEncodeHashKey(key, args[i])

		if v, err = db.db.Get(ek); err != nil {
			return 0, err
		} else if v == nil {
			continue
		} else {
			num++
			t.Delete(ek)
		}
	}

	if _, err = db.hIncrSize(key, -num); err != nil {
		return 0, err
	}

	err = t.Commit()

	return num, err
}

func (db *DB) hIncrSize(key []byte, delta int64) (int64, error) {
	t := db.hashTx
	sk := db.hEncodeSizeKey(key)

	var err error
	var size int64 = 0
	if size, err = Int64(db.db.Get(sk)); err != nil {
		return 0, err
	} else {
		size += delta
		if size <= 0 {
			size = 0
			t.Delete(sk)
			db.rmExpire(t, hExpType, key)
		} else {
			t.Put(sk, PutInt64(size))
		}
	}

	return size, nil
}

func (db *DB) HIncrBy(key []byte, field []byte, delta int64) (int64, error) {
	if err := checkHashKFSize(key, field); err != nil {
		return 0, err
	}

	t := db.hashTx
	var ek []byte
	var err error

	t.Lock()
	defer t.Unlock()

	ek = db.hEncodeHashKey(key, field)

	var n int64 = 0
	if n, err = StrInt64(db.db.Get(ek)); err != nil {
		return 0, err
	}

	n += delta

	_, err = db.hSetItem(key, field, StrPutInt64(n))
	if err != nil {
		return 0, err
	}

	err = t.Commit()

	return n, err
}

func (db *DB) HGetAll(key []byte) ([]interface{}, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	start := db.hEncodeStartKey(key)
	stop := db.hEncodeStopKey(key)

	v := make([]interface{}, 0, 16)

	it := db.db.RangeLimitIterator(start, stop, leveldb.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		_, k, err := db.hDecodeHashKey(it.Key())
		if err != nil {
			return nil, err
		}
		v = append(v, k)
		v = append(v, it.Value())
	}

	it.Close()

	return v, nil
}

func (db *DB) HKeys(key []byte) ([]interface{}, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	start := db.hEncodeStartKey(key)
	stop := db.hEncodeStopKey(key)

	v := make([]interface{}, 0, 16)

	it := db.db.RangeLimitIterator(start, stop, leveldb.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		_, k, err := db.hDecodeHashKey(it.Key())
		if err != nil {
			return nil, err
		}
		v = append(v, k)
	}

	it.Close()

	return v, nil
}

func (db *DB) HValues(key []byte) ([]interface{}, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	start := db.hEncodeStartKey(key)
	stop := db.hEncodeStopKey(key)

	v := make([]interface{}, 0, 16)

	it := db.db.RangeLimitIterator(start, stop, leveldb.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		v = append(v, it.Value())
	}

	it.Close()

	return v, nil
}

func (db *DB) HClear(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	t := db.hashTx
	t.Lock()
	defer t.Unlock()

	num := db.hDelete(t, key)
	db.rmExpire(t, hExpType, key)

	err := t.Commit()
	return num, err
}

func (db *DB) hFlush() (drop int64, err error) {
	t := db.kvTx
	t.Lock()
	defer t.Unlock()

	minKey := make([]byte, 2)
	minKey[0] = db.index
	minKey[1] = hashType

	maxKey := make([]byte, 2)
	maxKey[0] = db.index
	maxKey[1] = hSizeType + 1

	it := db.db.RangeLimitIterator(minKey, maxKey, leveldb.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		drop++
		if drop&1023 == 0 {
			if err = t.Commit(); err != nil {
				return
			}
		}
	}
	it.Close()

	db.expFlush(t, hExpType)

	err = t.Commit()
	return
}

func (db *DB) HScan(key []byte, field []byte, count int, inclusive bool) ([]FVPair, error) {
	var minKey []byte
	if field != nil {
		if err := checkHashKFSize(key, field); err != nil {
			return nil, err
		}
		minKey = db.hEncodeHashKey(key, field)
	} else {
		minKey = db.hEncodeStartKey(key)
	}

	maxKey := db.hEncodeStopKey(key)

	if count <= 0 {
		count = defaultScanCount
	}

	v := make([]FVPair, 0, 2*count)

	rangeType := leveldb.RangeROpen
	if !inclusive {
		rangeType = leveldb.RangeOpen
	}

	it := db.db.RangeLimitIterator(minKey, maxKey, rangeType, 0, count)
	for ; it.Valid(); it.Next() {
		if _, f, err := db.hDecodeHashKey(it.Key()); err != nil {
			continue
		} else {
			v = append(v, FVPair{Field: f, Value: it.Value()})
		}
	}
	it.Close()

	return v, nil
}

func (db *DB) HExpire(key []byte, duration int64) (int64, error) {
	if duration <= 0 {
		return 0, errExpireValue
	}

	return db.hExpireAt(key, time.Now().Unix()+duration)
}

func (db *DB) HExpireAt(key []byte, when int64) (int64, error) {
	if when <= time.Now().Unix() {
		return 0, errExpireValue
	}

	return db.hExpireAt(key, when)
}

func (db *DB) HTTL(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return -1, err
	}

	return db.ttl(hExpType, key)
}
