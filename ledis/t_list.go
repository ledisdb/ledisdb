package ledis

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/go-leveldb/leveldb"
)

const (
	listHeadSeq int32 = 1
	listTailSeq int32 = 2

	listMinSeq     int32 = 1000
	listMaxSeq     int32 = 1<<31 - 1000
	listInitialSeq int32 = listMinSeq + (listMaxSeq-listMinSeq)/2
)

var errLMetaKey = errors.New("invalid lmeta key")
var errListKey = errors.New("invalid list key")
var errListSeq = errors.New("invalid list sequence, overflow")

func (db *DB) lEncodeMetaKey(key []byte) []byte {
	buf := make([]byte, len(key)+2)
	buf[0] = db.index
	buf[1] = lMetaType

	copy(buf[2:], key)
	return buf
}

func (db *DB) lDecodeMetaKey(ek []byte) ([]byte, error) {
	if len(ek) < 2 || ek[0] != db.index || ek[1] != lMetaType {
		return nil, errLMetaKey
	}

	return ek[2:], nil
}

func (db *DB) lEncodeListKey(key []byte, seq int32) []byte {
	buf := make([]byte, len(key)+8)

	pos := 0
	buf[pos] = db.index
	pos++
	buf[pos] = listType
	pos++

	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2

	copy(buf[pos:], key)
	pos += len(key)

	binary.BigEndian.PutUint32(buf[pos:], uint32(seq))

	return buf
}

func (db *DB) lDecodeListKey(ek []byte) (key []byte, seq int32, err error) {
	if len(ek) < 8 || ek[0] != db.index || ek[1] != listType {
		err = errListKey
		return
	}

	keyLen := int(binary.BigEndian.Uint16(ek[2:]))
	if keyLen+8 != len(ek) {
		err = errListKey
		return
	}

	key = ek[4 : 4+keyLen]
	seq = int32(binary.BigEndian.Uint32(ek[4+keyLen:]))
	return
}

func (db *DB) lpush(key []byte, whereSeq int32, args ...[]byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	var headSeq int32
	var tailSeq int32
	var size int32
	var err error

	metaKey := db.lEncodeMetaKey(key)

	if len(args) == 0 {
		_, _, size, err := db.lGetMeta(metaKey)
		return int64(size), err
	}

	t := db.listTx
	t.Lock()
	defer t.Unlock()

	if headSeq, tailSeq, size, err = db.lGetMeta(metaKey); err != nil {
		return 0, err
	}

	var delta int32 = 1
	var seq int32 = 0
	if whereSeq == listHeadSeq {
		delta = -1
		seq = headSeq
	} else {
		seq = tailSeq
	}

	if size == 0 {
		headSeq = listInitialSeq
		tailSeq = listInitialSeq
		seq = headSeq
	} else {
		seq += delta
	}

	for i := 0; i < len(args); i++ {
		ek := db.lEncodeListKey(key, seq+int32(i)*delta)
		t.Put(ek, args[i])
		//to do add binlog
	}

	seq += int32(len(args)-1) * delta

	if seq <= listMinSeq || seq >= listMaxSeq {
		return 0, errListSeq
	}

	size += int32(len(args))

	if whereSeq == listHeadSeq {
		headSeq = seq
	} else {
		tailSeq = seq
	}

	db.lSetMeta(metaKey, headSeq, tailSeq, size)

	err = t.Commit()

	return int64(size), err
}

func (db *DB) lpop(key []byte, whereSeq int32) ([]byte, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	t := db.listTx
	t.Lock()
	defer t.Unlock()

	var headSeq int32
	var tailSeq int32
	var size int32
	var err error

	metaKey := db.lEncodeMetaKey(key)

	headSeq, tailSeq, size, err = db.lGetMeta(metaKey)

	if err != nil {
		return nil, err
	}

	var seq int32 = 0
	var delta int32 = 1
	if whereSeq == listHeadSeq {
		seq = headSeq
	} else {
		delta = -1
		seq = tailSeq
	}

	itemKey := db.lEncodeListKey(key, seq)
	var value []byte
	value, err = db.db.Get(itemKey)
	if err != nil {
		return nil, err
	}

	t.Delete(itemKey)
	seq += delta

	size--
	if size <= 0 {
		t.Delete(metaKey)
	} else {
		if whereSeq == listHeadSeq {
			headSeq = seq
		} else {
			tailSeq = seq
		}

		db.lSetMeta(metaKey, headSeq, tailSeq, size)
	}

	//todo add binlog
	err = t.Commit()
	return value, err
}

func (db *DB) lGetSeq(key []byte, whereSeq int32) (int64, error) {
	ek := db.lEncodeListKey(key, whereSeq)

	return Int64(db.db.Get(ek))
}

func (db *DB) lGetMeta(ek []byte) (headSeq int32, tailSeq int32, size int32, err error) {
	var v []byte
	v, err = db.db.Get(ek)
	if err != nil {
		return
	} else if v == nil {
		size = 0
		return
	} else {
		headSeq = int32(binary.LittleEndian.Uint32(v[0:4]))
		tailSeq = int32(binary.LittleEndian.Uint32(v[4:8]))
		size = int32(binary.LittleEndian.Uint32(v[8:]))
	}
	return
}

func (db *DB) lSetMeta(ek []byte, headSeq int32, tailSeq int32, size int32) {
	t := db.listTx

	buf := make([]byte, 12)

	binary.LittleEndian.PutUint32(buf[0:4], uint32(headSeq))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(tailSeq))
	binary.LittleEndian.PutUint32(buf[8:], uint32(size))

	t.Put(ek, buf)
}

func (db *DB) LIndex(key []byte, index int32) ([]byte, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	var seq int32
	var headSeq int32
	var tailSeq int32
	var err error

	metaKey := db.lEncodeMetaKey(key)

	headSeq, tailSeq, _, err = db.lGetMeta(metaKey)
	if err != nil {
		return nil, err
	}

	if index >= 0 {
		seq = headSeq + index
	} else {
		seq = tailSeq + index + 1
	}

	sk := db.lEncodeListKey(key, seq)
	return db.db.Get(sk)
}

func (db *DB) LLen(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	ek := db.lEncodeMetaKey(key)
	_, _, size, err := db.lGetMeta(ek)
	return int64(size), err
}

func (db *DB) LPop(key []byte) ([]byte, error) {
	return db.lpop(key, listHeadSeq)
}

func (db *DB) LPush(key []byte, args ...[]byte) (int64, error) {
	return db.lpush(key, listHeadSeq, args...)
}

func (db *DB) LRange(key []byte, start int32, stop int32) ([]interface{}, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	v := make([]interface{}, 0, 16)

	var startSeq int32
	var stopSeq int32

	if start > stop {
		return []interface{}{}, nil
	}

	var headSeq int32
	var tailSeq int32
	var err error

	metaKey := db.lEncodeMetaKey(key)

	if headSeq, tailSeq, _, err = db.lGetMeta(metaKey); err != nil {
		return nil, err
	}

	if start >= 0 && stop >= 0 {
		startSeq = headSeq + start
		stopSeq = headSeq + stop
	} else if start < 0 && stop < 0 {
		startSeq = tailSeq + start + 1
		stopSeq = tailSeq + stop + 1
	} else {
		//start < 0 && stop > 0
		startSeq = tailSeq + start + 1
		stopSeq = headSeq + stop
	}

	if startSeq < listMinSeq {
		startSeq = listMinSeq
	} else if stopSeq > listMaxSeq {
		stopSeq = listMaxSeq
	}

	startKey := db.lEncodeListKey(key, startSeq)
	stopKey := db.lEncodeListKey(key, stopSeq)
	it := db.db.Iterator(startKey, stopKey, leveldb.RangeClose, 0, -1)
	for ; it.Valid(); it.Next() {
		v = append(v, it.Value())
	}

	it.Close()

	return v, nil
}

func (db *DB) RPop(key []byte) ([]byte, error) {
	return db.lpop(key, listTailSeq)
}

func (db *DB) RPush(key []byte, args ...[]byte) (int64, error) {
	return db.lpush(key, listTailSeq, args...)
}

func (db *DB) LClear(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	mk := db.lEncodeMetaKey(key)

	t := db.listTx
	t.Lock()
	defer t.Unlock()

	var headSeq int32
	var tailSeq int32
	var err error

	headSeq, tailSeq, _, err = db.lGetMeta(mk)

	if err != nil {
		return 0, err
	}

	var num int64 = 0
	startKey := db.lEncodeListKey(key, headSeq)
	stopKey := db.lEncodeListKey(key, tailSeq)

	it := db.db.Iterator(startKey, stopKey, leveldb.RangeClose, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		num++
	}

	it.Close()

	t.Delete(mk)

	err = t.Commit()
	return num, err
}
