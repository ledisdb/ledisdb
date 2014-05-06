package ssdb

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/golib/hack"
	"strconv"
)

var errHashKey = errors.New("invalid hash key")
var errHSizeKey = errors.New("invalid hsize key")

const (
	hashStartSep byte = ':'
	hashStopSep  byte = ';'
)

func encode_hsize_key(key []byte) []byte {
	buf := make([]byte, len(key)+1)
	buf[0] = HSIZE_TYPE

	copy(buf[1:], key)
	return buf
}

func decode_hsize_key(ek []byte) ([]byte, error) {
	if len(ek) == 0 || ek[0] != HSIZE_TYPE {
		return nil, errHSizeKey
	}

	return ek[1:], nil
}

func encode_hash_key(key []byte, field []byte) []byte {
	buf := make([]byte, len(key)+len(field)+1+4+1)

	pos := 0
	buf[pos] = HASH_TYPE
	pos++
	binary.BigEndian.PutUint32(buf[pos:], uint32(len(key)))
	pos += 4

	copy(buf[pos:], key)
	pos += len(key)

	buf[pos] = hashStartSep
	pos++
	copy(buf[pos:], field)

	return buf
}

func encode_hash_start_key(key []byte) []byte {
	buf := make([]byte, len(key)+1+4+1)

	pos := 0
	buf[pos] = HASH_TYPE
	pos++
	binary.BigEndian.PutUint32(buf[pos:], uint32(len(key)))
	pos += 4

	copy(buf[pos:], key)
	pos += len(key)

	buf[pos] = hashStartSep
	return buf
}

func encode_hash_stop_key(key []byte) []byte {
	buf := make([]byte, len(key)+1+4+1)

	pos := 0
	buf[pos] = HASH_TYPE
	pos++
	binary.BigEndian.PutUint32(buf[pos:], uint32(len(key)))
	pos += 4

	copy(buf[pos:], key)
	pos += len(key)

	buf[pos] = hashStopSep
	return buf
}

func decode_hash_key(ek []byte) ([]byte, []byte, error) {
	if len(ek) < 6 || ek[0] != HASH_TYPE {
		return nil, nil, errHashKey
	}

	pos := 1
	keyLen := int(binary.BigEndian.Uint32(ek[pos:]))
	pos += 4

	if keyLen+6 > len(ek) {
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

func (a *App) hash_len(key []byte) (int64, error) {
	return a.db.GetInt(encode_hsize_key(key))
}

func (a *App) hash_setItem(key []byte, field []byte, value []byte) (int64, error) {
	t := a.hashTx

	ek := encode_hash_key(key, field)
	sk := encode_hsize_key(key)

	size, err := a.db.GetInt(sk)
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	if v, _ := a.db.Get(ek); v != nil {
		n = 0
	} else {
		size++
		t.Put(sk, hack.Slice(strconv.FormatInt(size, 10)))
	}

	t.Put(ek, value)
	return n, nil
}

func (a *App) hash_set(key []byte, field []byte, value []byte) (int64, error) {
	t := a.hashTx
	t.Lock()
	defer t.Unlock()

	n, err := a.hash_setItem(key, field, value)
	if err != nil {
		return 0, err
	}

	//todo add binlog

	err = t.Commit()
	return n, err
}

func (a *App) hash_get(key []byte, field []byte) ([]byte, error) {
	return a.db.Get(encode_hash_key(key, field))
}

func (a *App) hash_mset(key []byte, args [][]byte) error {
	sk := encode_hsize_key(key)

	t := a.hashTx
	t.Lock()
	defer t.Unlock()

	size, err := a.db.GetInt(sk)
	if err != nil {
		return err
	}

	for i := 0; i < len(args); i += 2 {
		ek := encode_hash_key(key, args[i])
		if v, _ := a.db.Get(ek); v == nil {
			size++
		}

		t.Put(ek, args[i+1])
	}

	t.Put(sk, hack.Slice(strconv.FormatInt(size, 10)))

	//todo add binglog
	err = t.Commit()
	return err
}

func (a *App) hash_mget(key []byte, args [][]byte) ([]interface{}, error) {
	r := make([]interface{}, len(args))
	for i := 0; i < len(args); i++ {
		v, err := a.db.Get(encode_hash_key(key, args[i]))
		if err != nil {
			return nil, err
		}

		r[i] = v
	}

	return r, nil
}

func (a *App) hash_del(key []byte, args [][]byte) (int64, error) {
	sk := encode_hsize_key(key)

	t := a.hashTx
	t.Lock()
	defer t.Unlock()

	size, err := a.db.GetInt(sk)
	if err != nil {
		return 0, err
	}

	var num int64 = 0
	for i := 0; i < len(args); i++ {
		ek := encode_hash_key(key, args[i])
		if v, err := a.db.Get(ek); err != nil {
			return 0, err
		} else if v == nil {
			continue
		} else {
			num++
			size--
			t.Delete(ek)
		}
	}

	if size <= 0 {
		t.Delete(sk)
	} else {
		t.Put(sk, hack.Slice(strconv.FormatInt(size, 10)))
	}

	err = t.Commit()

	return num, err
}

func (a *App) hash_incrby(key []byte, field []byte, delta int64) (int64, error) {
	t := a.hashTx
	t.Lock()
	defer t.Unlock()

	ek := encode_hash_key(key, field)

	var n int64 = 0
	v, err := a.db.Get(ek)
	if err != nil {
		return 0, err
	} else if v != nil {
		if n, err = strconv.ParseInt(hack.String(v), 10, 64); err != nil {
			return 0, err
		}
	}

	n += delta

	_, err = a.hash_setItem(key, field, hack.Slice(strconv.FormatInt(n, 10)))
	if err != nil {
		return 0, err
	}

	err = t.Commit()

	return n, err
}

func (a *App) hash_getall(key []byte) ([]interface{}, error) {
	start := encode_hash_start_key(key)
	stop := encode_hash_stop_key(key)

	v := make([]interface{}, 0, 16)

	it := a.db.Iterator(start, stop, 0)
	for ; it.Valid(); it.Next() {
		_, k, err := decode_hash_key(it.Key())
		if err != nil {
			return nil, err
		}
		v = append(v, k)
		v = append(v, it.Value())
	}

	return v, nil
}

func (a *App) hash_keys(key []byte) ([]interface{}, error) {
	start := encode_hash_start_key(key)
	stop := encode_hash_stop_key(key)

	v := make([]interface{}, 0, 16)

	it := a.db.Iterator(start, stop, 0)
	for ; it.Valid(); it.Next() {
		_, k, err := decode_hash_key(it.Key())
		if err != nil {
			return nil, err
		}
		v = append(v, k)
	}

	return v, nil
}

func (a *App) hash_values(key []byte) ([]interface{}, error) {
	start := encode_hash_start_key(key)
	stop := encode_hash_stop_key(key)

	v := make([]interface{}, 0, 16)

	it := a.db.Iterator(start, stop, 0)
	for ; it.Valid(); it.Next() {
		v = append(v, it.Value())
	}

	return v, nil
}
