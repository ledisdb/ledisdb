package ssdb

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/golib/hack"
	"strconv"
)

const (
	listHeadSeq int64 = 1
	listTailSeq int64 = 2

	listMinSeq     int64 = 1000
	listMaxSeq     int64 = 1<<63 - 1000
	listInitialSeq int64 = listMinSeq + (listMaxSeq-listMinSeq)/2
)

var errLSizeKey = errors.New("invalid lsize key")
var errListKey = errors.New("invalid list key")
var errListSeq = errors.New("invalid list sequence, overflow")

func encode_lsize_key(key []byte) []byte {
	buf := make([]byte, len(key)+1)
	buf[0] = LSIZE_TYPE

	copy(buf[1:], key)
	return buf
}

func decode_lsize_key(ek []byte) ([]byte, error) {
	if len(ek) == 0 || ek[0] != LSIZE_TYPE {
		return nil, errLSizeKey
	}

	return ek[1:], nil
}

func encode_list_key(key []byte, seq int64) []byte {
	buf := make([]byte, len(key)+13)

	pos := 0
	buf[pos] = LIST_TYPE
	pos++

	binary.BigEndian.PutUint32(buf[pos:], uint32(len(key)))
	pos += 4

	copy(buf[pos:], key)
	pos += len(key)

	binary.BigEndian.PutUint64(buf[pos:], uint64(seq))

	return buf
}

func decode_list_key(ek []byte) (key []byte, seq int64, err error) {
	if len(ek) < 13 || ek[0] != LIST_TYPE {
		err = errListKey
		return
	}

	keyLen := int(binary.BigEndian.Uint32(ek[1:]))
	if keyLen+13 != len(ek) {
		err = errListKey
		return
	}

	key = ek[5 : 5+keyLen]
	seq = int64(binary.BigEndian.Uint64(ek[5+keyLen:]))
	return
}

func (a *App) list_lpush(key []byte, args [][]byte) (int64, error) {
	return a.list_push(key, args, listHeadSeq)
}

func (a *App) list_rpush(key []byte, args [][]byte) (int64, error) {
	return a.list_push(key, args, listTailSeq)
}

func (a *App) list_lpop(key []byte) ([]byte, error) {
	return a.list_pop(key, listHeadSeq)
}

func (a *App) list_rpop(key []byte) ([]byte, error) {
	return a.list_pop(key, listTailSeq)
}

func (a *App) list_getSeq(key []byte, whereSeq int64) (int64, error) {
	ek := encode_list_key(key, whereSeq)

	return a.db.GetInt(ek)
}

func (a *App) list_len(key []byte) (int64, error) {
	ek := encode_lsize_key(key)

	return a.db.GetInt(ek)
}

func (a *App) list_push(key []byte, args [][]byte, whereSeq int64) (int64, error) {
	t := a.listTx
	t.Lock()
	defer t.Unlock()

	seq, err := a.list_getSeq(key, whereSeq)
	if err != nil {
		return 0, err
	}

	var size int64 = 0

	var delta int64 = 1
	if whereSeq == listHeadSeq {
		delta = -1
	}

	if seq == 0 {
		seq = listInitialSeq

		t.Put(encode_list_key(key, listHeadSeq), hack.Slice(strconv.FormatInt(seq, 10)))
		t.Put(encode_list_key(key, listTailSeq), hack.Slice(strconv.FormatInt(seq, 10)))
	} else {
		size, err = a.list_len(key)
		if err != nil {
			return 0, err
		}

		seq += delta
	}

	for i := 0; i < len(args); i++ {
		t.Put(encode_list_key(key, seq+int64(i)*delta), args[i])
		//to do add binlog
	}

	seq += int64(len(args)-1) * delta

	if seq <= listMinSeq || seq >= listMaxSeq {
		return 0, errListSeq
	}

	size += int64(len(args))

	t.Put(encode_lsize_key(key), hack.Slice(strconv.FormatInt(size, 10)))
	t.Put(encode_list_key(key, whereSeq), hack.Slice(strconv.FormatInt(seq, 10)))

	err = t.Commit()

	return size, err
}

func (a *App) list_pop(key []byte, whereSeq int64) ([]byte, error) {
	t := a.listTx
	t.Lock()
	defer t.Unlock()

	var delta int64 = 1
	if whereSeq == listTailSeq {
		delta = -1
	}

	seq, err := a.list_getSeq(key, whereSeq)
	if err != nil {
		return nil, err
	}

	var value []byte
	value, err = a.db.Get(encode_list_key(key, seq))
	if err != nil {
		return nil, err
	}

	t.Delete(encode_list_key(key, seq))
	seq += delta

	var size int64
	size, err = a.list_len(key)
	if err != nil {
		return nil, err
	}

	size--
	if size <= 0 {
		t.Delete(encode_lsize_key(key))
		t.Delete(encode_list_key(key, listHeadSeq))
		t.Delete(encode_list_key(key, listTailSeq))
	} else {
		t.Put(encode_list_key(key, whereSeq), hack.Slice(strconv.FormatInt(seq, 10)))
		t.Put(encode_lsize_key(key), hack.Slice(strconv.FormatInt(size, 10)))
	}

	//todo add binlog
	err = t.Commit()
	return value, err
}

func (a *App) list_range(key []byte, start int64, stop int64) ([]interface{}, error) {
	v := make([]interface{}, 0, 16)

	var startSeq int64
	var stopSeq int64

	if start > stop {
		return []interface{}{}, nil
	} else if start >= 0 && stop >= 0 {
		seq, err := a.list_getSeq(key, listHeadSeq)
		if err != nil {
			return nil, err
		}

		startSeq = seq + start
		stopSeq = seq + stop + 1

	} else if start < 0 && stop < 0 {
		seq, err := a.list_getSeq(key, listTailSeq)
		if err != nil {
			return nil, err
		}

		startSeq = seq + start + 1
		stopSeq = seq + stop + 2
	} else {
		//start < 0 && stop > 0
		var err error
		startSeq, err = a.list_getSeq(key, listTailSeq)
		if err != nil {
			return nil, err
		}

		startSeq += start + 1

		stopSeq, err = a.list_getSeq(key, listHeadSeq)
		if err != nil {
			return nil, err
		}

		stopSeq += stop + 1
	}

	if startSeq < listMinSeq {
		startSeq = listMinSeq
	} else if stopSeq > listMaxSeq {
		stopSeq = listMaxSeq
	}

	it := a.db.Iterator(encode_list_key(key, startSeq),
		encode_list_key(key, stopSeq), 0)
	for ; it.Valid(); it.Next() {
		v = append(v, it.Value())
	}

	it.Close()

	return v, nil
}

func (a *App) list_index(key []byte, index int64) ([]byte, error) {
	var seq int64
	var err error
	if index >= 0 {
		seq, err = a.list_getSeq(key, listHeadSeq)
		if err != nil {
			return nil, err
		}

		seq = seq + index

	} else {
		seq, err = a.list_getSeq(key, listTailSeq)
		if err != nil {
			return nil, err
		}

		seq = seq + index + 1
	}

	return a.db.Get(encode_list_key(key, seq))
}
