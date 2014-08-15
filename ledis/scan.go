package ledis

import (
	"errors"
	"github.com/siddontang/ledisdb/store"
)

var errDataType = errors.New("error data type")
var errMetaKey = errors.New("error meta key")

func (db *DB) scan(dataType byte, key []byte, count int, inclusive bool) ([][]byte, error) {
	var minKey, maxKey []byte
	var err error
	if key != nil {
		if err = checkKeySize(key); err != nil {
			return nil, err
		}
		if minKey, err = db.encodeMetaKey(dataType, key); err != nil {
			return nil, err
		}

	} else {
		if minKey, err = db.encodeMinKey(dataType); err != nil {
			return nil, err
		}
	}

	if maxKey, err = db.encodeMaxKey(dataType); err != nil {
		return nil, err
	}

	if count <= 0 {
		count = defaultScanCount
	}

	v := make([][]byte, 0, count)

	rangeType := store.RangeROpen
	if !inclusive {
		rangeType = store.RangeOpen
	}

	it := db.db.RangeLimitIterator(minKey, maxKey, rangeType, 0, count)

	for ; it.Valid(); it.Next() {
		if k, err := db.decodeMetaKey(dataType, it.Key()); err != nil {
			continue
		} else {
			v = append(v, k)
		}
	}
	it.Close()
	return v, nil
}

func (db *DB) encodeMinKey(dataType byte) ([]byte, error) {
	return db.encodeMetaKey(dataType, nil)
}

func (db *DB) encodeMaxKey(dataType byte) ([]byte, error) {
	k, err := db.encodeMetaKey(dataType, nil)
	if err != nil {
		return nil, err
	}
	k[len(k)-1] = dataType + 1
	return k, nil
}

func (db *DB) encodeMetaKey(dataType byte, key []byte) ([]byte, error) {
	switch dataType {
	case KVType:
		return db.encodeKVKey(key), nil
	case LMetaType:
		return db.lEncodeMetaKey(key), nil
	case HSizeType:
		return db.hEncodeSizeKey(key), nil
	case ZSizeType:
		return db.zEncodeSizeKey(key), nil
	case BitMetaType:
		return db.bEncodeMetaKey(key), nil
	default:
		return nil, errDataType
	}
}
func (db *DB) decodeMetaKey(dataType byte, ek []byte) ([]byte, error) {
	if len(ek) < 2 || ek[0] != db.index || ek[1] != dataType {
		return nil, errMetaKey
	}
	return ek[2:], nil
}
