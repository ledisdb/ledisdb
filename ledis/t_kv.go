package ledis

import (
	"errors"
	"github.com/siddontang/go-leveldb/leveldb"
)

type KVPair struct {
	Key   []byte
	Value []byte
}

var errKVKey = errors.New("invalid encode kv key")

func checkKeySize(key []byte) error {
	if len(key) > MaxKeySize {
		return ErrKeySize
	}
	return nil
}

func (db *DB) encodeKVKey(key []byte) []byte {
	ek := make([]byte, len(key)+2)
	ek[0] = db.index
	ek[1] = kvType
	copy(ek[2:], key)
	return ek
}

func (db *DB) decodeKVKey(ek []byte) ([]byte, error) {
	if len(ek) < 2 || ek[0] != db.index || ek[1] != kvType {
		return nil, errKVKey
	}

	return ek[2:], nil
}

func (db *DB) encodeKVMinKey() []byte {
	ek := db.encodeKVKey(nil)
	return ek
}

func (db *DB) encodeKVMaxKey() []byte {
	ek := db.encodeKVKey(nil)
	ek[len(ek)-1] = kvType + 1
	return ek
}

func (db *DB) incr(key []byte, delta int64) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	var err error
	key = db.encodeKVKey(key)

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

func (db *DB) Del(keys ...[]byte) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var err error
	for i := range keys {
		keys[i] = db.encodeKVKey(keys[i])
	}

	t := db.kvTx

	t.Lock()
	defer t.Unlock()

	for i := range keys {
		t.Delete(keys[i])
		//todo binlog
	}

	err = t.Commit()
	return int64(len(keys)), err
}

func (db *DB) Exists(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	var err error
	key = db.encodeKVKey(key)

	var v []byte
	v, err = db.db.Get(key)
	if v != nil && err == nil {
		return 1, nil
	}

	return 0, err
}

func (db *DB) Get(key []byte) ([]byte, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	key = db.encodeKVKey(key)

	return db.db.Get(key)
}

func (db *DB) GetSet(key []byte, value []byte) ([]byte, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	key = db.encodeKVKey(key)

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

func (db *DB) MGet(keys ...[]byte) ([]interface{}, error) {
	values := make([]interface{}, len(keys))

	var err error
	var value []byte
	for i := range keys {
		if err := checkKeySize(keys[i]); err != nil {
			return nil, err
		}

		if value, err = db.db.Get(db.encodeKVKey(keys[i])); err != nil {
			return nil, err
		}

		values[i] = value
	}

	return values, nil
}

func (db *DB) MSet(args ...KVPair) error {
	if len(args) == 0 {
		return nil
	}

	t := db.kvTx

	var err error
	var key []byte
	var value []byte

	t.Lock()
	defer t.Unlock()

	for i := 0; i < len(args); i++ {
		if err := checkKeySize(args[i].Key); err != nil {
			return err
		}

		key = db.encodeKVKey(args[i].Key)

		value = args[i].Value

		t.Put(key, value)

		//todo binlog
	}

	err = t.Commit()
	return err
}

func (db *DB) Set(key []byte, value []byte) error {
	if err := checkKeySize(key); err != nil {
		return err
	}

	var err error
	key = db.encodeKVKey(key)

	t := db.kvTx

	t.Lock()
	defer t.Unlock()

	t.Put(key, value)

	//todo, binlog

	err = t.Commit()

	return err
}

func (db *DB) SetNX(key []byte, value []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	var err error
	key = db.encodeKVKey(key)

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

func (db *DB) KvFlush() (drop int64, err error) {
	t := db.kvTx
	t.Lock()
	defer t.Unlock()

	minKey := db.encodeKVMinKey()
	maxKey := db.encodeKVMaxKey()

	it := db.db.Iterator(minKey, maxKey, leveldb.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		drop++
	}

	err = t.Commit()
	return
}

func (db *DB) Scan(cursor int, count int) ([]interface{}, error) {
	minKey := db.encodeKVMinKey()
	maxKey := db.encodeKVMaxKey()

	if count <= 0 {
		count = defaultScanCount
	}

	v := make([]interface{}, 2)
	r := make([]interface{}, 0, count)

	var num int = 0
	it := db.db.Iterator(minKey, maxKey, leveldb.RangeROpen, cursor, count)
	for ; it.Valid(); it.Next() {
		num++

		if key, err := db.decodeKVKey(it.Key()); err != nil {
			continue
		} else {
			r = append(r, key)
		}
	}

	if num < count {
		v[0] = int64(0)
	} else {
		v[0] = int64(cursor + count)
	}

	v[1] = r

	return v, nil
}
