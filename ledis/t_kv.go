package ledis

import (
	"errors"
)

var errKVKey = errors.New("invalid encode kv key")

func encode_kv_key(key []byte) []byte {
	ek := make([]byte, len(key)+1)
	ek[0] = KV_TYPE
	copy(ek[1:], key)
	return ek
}

func decode_kv_key(ek []byte) ([]byte, error) {
	if len(ek) == 0 || ek[0] != KV_TYPE {
		return nil, errKVKey
	}

	return ek[1:], nil
}

func (db *DB) incr(key []byte, delta int64) (int64, error) {
	key = encode_kv_key(key)
	var err error

	t := db.kvTx

	t.Lock()
	defer t.Unlock()

	var n int64
	n, err = StrInt64(db.db.Get(key))
	if err != nil {
		return 0, err
	}

	n += delta

	t.Put(key, StrPutInt64(n))

	//todo binlog

	err = t.Commit()
	return n, err
}

func (db *DB) Decr(key []byte) (int64, error) {
	return db.incr(key, -1)
}

func (db *DB) DecrBy(key []byte, decrement int64) (int64, error) {
	return db.incr(key, -decrement)
}

func (db *DB) Del(keys [][]byte) (int64, error) {
	for i := range keys {
		keys[i] = encode_kv_key(keys[i])
	}

	t := db.kvTx

	t.Lock()
	defer t.Unlock()

	for i := range keys {
		t.Delete(keys[i])
		//todo binlog
	}

	err := t.Commit()
	return int64(len(keys)), err
}

func (db *DB) Exists(key []byte) (int64, error) {
	key = encode_kv_key(key)
	var err error

	var v []byte
	v, err = db.db.Get(key)
	if v != nil && err == nil {
		return 1, nil
	}

	return 0, err
}

func (db *DB) Get(key []byte) ([]byte, error) {
	key = encode_kv_key(key)

	return db.db.Get(key)
}

func (db *DB) GetSet(key []byte, value []byte) ([]byte, error) {
	key = encode_kv_key(key)

	t := db.kvTx

	t.Lock()
	defer t.Unlock()

	oldValue, err := db.db.Get(key)
	if err != nil {
		return nil, err
	}

	t.Put(key, value)
	//todo, binlog

	err = t.Commit()

	return oldValue, err
}

func (db *DB) Incr(key []byte) (int64, error) {
	return db.incr(key, 1)
}

func (db *DB) IncryBy(key []byte, increment int64) (int64, error) {
	return db.incr(key, increment)
}

func (db *DB) MGet(keys [][]byte) ([]interface{}, error) {
	values := make([]interface{}, len(keys))

	for i := range keys {
		key := encode_kv_key(keys[i])
		value, err := db.db.Get(key)
		if err != nil {
			return nil, err
		}

		values[i] = value
	}

	return values, nil
}

func (db *DB) MSet(args [][]byte) error {
	t := db.kvTx

	t.Lock()
	defer t.Unlock()

	for i := 0; i < len(args); i += 2 {
		key := encode_kv_key(args[i])
		value := args[i+1]

		t.Put(key, value)

		//todo binlog
	}

	err := t.Commit()
	return err
}

func (db *DB) Set(key []byte, value []byte) error {
	key = encode_kv_key(key)
	var err error

	t := db.kvTx

	t.Lock()
	defer t.Unlock()

	t.Put(key, value)

	//todo, binlog

	err = t.Commit()

	return err
}

func (db *DB) SetNX(key []byte, value []byte) (int64, error) {
	key = encode_kv_key(key)
	var err error

	var n int64 = 1

	t := db.kvTx

	t.Lock()
	defer t.Unlock()

	if v, err := db.db.Get(key); err != nil {
		return 0, err
	} else if v != nil {
		n = 0
	} else {
		t.Put(key, value)

		//todo binlog

		err = t.Commit()
	}

	return n, err
}
