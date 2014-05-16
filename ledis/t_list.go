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

func encode_lmeta_key(key []byte) []byte {
	buf := make([]byte, len(key)+1)
	buf[0] = LMETA_TYPE

	copy(buf[1:], key)
	return buf
}

func decode_lmeta_key(ek []byte) ([]byte, error) {
	if len(ek) == 0 || ek[0] != LMETA_TYPE {
		return nil, errLMetaKey
	}

	return ek[1:], nil
}

func encode_list_key(key []byte, seq int32) []byte {
	buf := make([]byte, len(key)+9)

	pos := 0
	buf[pos] = LIST_TYPE
	pos++

	binary.BigEndian.PutUint32(buf[pos:], uint32(len(key)))
	pos += 4

	copy(buf[pos:], key)
	pos += len(key)

	binary.BigEndian.PutUint32(buf[pos:], uint32(seq))

	return buf
}

func decode_list_key(ek []byte) (key []byte, seq int32, err error) {
	if len(ek) < 9 || ek[0] != LIST_TYPE {
		err = errListKey
		return
	}

	keyLen := int(binary.BigEndian.Uint32(ek[1:]))
	if keyLen+9 != len(ek) {
		err = errListKey
		return
	}

	key = ek[5 : 5+keyLen]
	seq = int32(binary.BigEndian.Uint32(ek[5+keyLen:]))
	return
}

func (db *DB) lpush(key []byte, whereSeq int32, args ...[]byte) (int64, error) {
	metaKey := encode_lmeta_key(key)

	if len(args) == 0 {
		_, _, size, err := db.lGetMeta(metaKey)
		return int64(size), err
	}

	t := db.listTx
	t.Lock()
	defer t.Unlock()

	headSeq, tailSeq, size, err := db.lGetMeta(metaKey)

	if err != nil {
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
		t.Put(encode_list_key(key, seq+int32(i)*delta), args[i])
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
	t := db.listTx
	t.Lock()
	defer t.Unlock()

	metaKey := encode_lmeta_key(key)
	headSeq, tailSeq, size, err := db.lGetMeta(metaKey)

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

	itemKey := encode_list_key(key, seq)
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
	ek := encode_list_key(key, whereSeq)

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
	var seq int32
	headSeq, tailSeq, _, err := db.lGetMeta(encode_lmeta_key(key))
	if err != nil {
		return nil, err
	}

	if index >= 0 {
		seq = headSeq + index
	} else {
		seq = tailSeq + index + 1
	}

	return db.db.Get(encode_list_key(key, seq))
}

func (db *DB) LLen(key []byte) (int64, error) {
	ek := encode_lmeta_key(key)
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
	v := make([]interface{}, 0, 16)

	var startSeq int32
	var stopSeq int32

	if start > stop {
		return []interface{}{}, nil
	}

	headSeq, tailSeq, _, err := db.lGetMeta(encode_lmeta_key(key))
	if err != nil {
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

	it := db.db.Iterator(encode_list_key(key, startSeq),
		encode_list_key(key, stopSeq), leveldb.RangeClose, 0, -1)
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
	mk := encode_lmeta_key(key)

	t := db.listTx
	t.Lock()
	defer t.Unlock()

	metaKey := encode_lmeta_key(key)
	headSeq, tailSeq, _, err := db.lGetMeta(metaKey)

	if err != nil {
		return 0, err
	}

	var num int64 = 0
	it := db.db.Iterator(encode_list_key(key, headSeq),
		encode_list_key(key, tailSeq), leveldb.RangeClose, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		num++
	}

	it.Close()

	t.Delete(mk)

	err = t.Commit()
	return num, err
}
