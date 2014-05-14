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

func (a *App) kv_get(key []byte) ([]byte, error) {
	key = encode_kv_key(key)

	return a.db.Get(key)
}

func (a *App) kv_set(key []byte, value []byte) error {
	key = encode_kv_key(key)
	var err error

	t := a.kvTx

	t.Lock()
	defer t.Unlock()

	t.Put(key, value)

	//todo, binlog

	err = t.Commit()

	return err
}

func (a *App) kv_getset(key []byte, value []byte) ([]byte, error) {
	key = encode_kv_key(key)

	t := a.kvTx

	t.Lock()
	defer t.Unlock()

	oldValue, err := a.db.Get(key)
	if err != nil {
		return nil, err
	}

	t.Put(key, value)
	//todo, binlog

	err = t.Commit()

	return oldValue, err
}

func (a *App) kv_setnx(key []byte, value []byte) (int64, error) {
	key = encode_kv_key(key)
	var err error

	var n int64 = 1

	t := a.kvTx

	t.Lock()
	defer t.Unlock()

	if v, err := a.db.Get(key); err != nil {
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

func (a *App) kv_exists(key []byte) (int64, error) {
	key = encode_kv_key(key)
	var err error

	var v []byte
	v, err = a.db.Get(key)
	if v != nil && err == nil {
		return 1, nil
	} else {
		return 0, err
	}
}

func (a *App) kv_incr(key []byte, delta int64) (int64, error) {
	key = encode_kv_key(key)
	var err error

	t := a.kvTx

	t.Lock()
	defer t.Unlock()

	var n int64
	n, err = StrInt64(a.db.Get(key))
	if err != nil {
		return 0, err
	}

	n += delta

	t.Put(key, StrPutInt64(n))

	//todo binlog

	err = t.Commit()
	return n, err
}

func (a *App) tx_del(keys [][]byte) (int64, error) {
	for i := range keys {
		keys[i] = encode_kv_key(keys[i])
	}

	t := a.kvTx

	t.Lock()
	defer t.Unlock()

	for i := range keys {
		t.Delete(keys[i])
		//todo binlog
	}

	err := t.Commit()
	return int64(len(keys)), err
}

func (a *App) tx_mset(args [][]byte) error {
	t := a.kvTx

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

func (a *App) kv_mget(args [][]byte) ([]interface{}, error) {
	values := make([]interface{}, len(args))

	for i := range args {
		key := encode_kv_key(args[i])
		value, err := a.db.Get(key)
		if err != nil {
			return nil, err
		}

		values[i] = value
	}

	return values, nil
}
