package ledis

import (
	"errors"
	"github.com/siddontang/ledisdb/store"
	"regexp"
)

var errDataType = errors.New("error data type")
var errMetaKey = errors.New("error meta key")

//fif inclusive is true, scan range [cursor, inf) else (cursor, inf)
func (db *DB) Scan(dataType DataType, cursor []byte, count int, inclusive bool, match string) ([][]byte, error) {
	storeDataType, err := getDataStoreType(dataType)
	if err != nil {
		return nil, err
	}

	return db.scanGeneric(storeDataType, cursor, count, inclusive, match, false)
}

//if inclusive is true, revscan range (-inf, cursor] else (inf, cursor)
func (db *DB) RevScan(dataType DataType, cursor []byte, count int, inclusive bool, match string) ([][]byte, error) {
	storeDataType, err := getDataStoreType(dataType)
	if err != nil {
		return nil, err
	}

	return db.scanGeneric(storeDataType, cursor, count, inclusive, match, true)
}

func getDataStoreType(dataType DataType) (byte, error) {
	var storeDataType byte
	switch dataType {
	case KV:
		storeDataType = KVType
	case LIST:
		storeDataType = LMetaType
	case HASH:
		storeDataType = HSizeType
	case SET:
		storeDataType = SSizeType
	case ZSET:
		storeDataType = ZSizeType
	default:
		return 0, errDataType
	}
	return storeDataType, nil
}

func buildMatchRegexp(match string) (*regexp.Regexp, error) {
	var err error
	var r *regexp.Regexp = nil

	if len(match) > 0 {
		if r, err = regexp.Compile(match); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (db *DB) scanGeneric(storeDataType byte, key []byte, count int,
	inclusive bool, match string, reverse bool) ([][]byte, error) {
	var minKey, maxKey []byte
	r, err := buildMatchRegexp(match)
	if err != nil {
		return nil, err
	}

	tp := store.RangeOpen

	if !reverse {
		if minKey, err = db.encodeScanMinKey(storeDataType, key); err != nil {
			return nil, err
		}
		if maxKey, err = db.encodeScanMaxKey(storeDataType, nil); err != nil {
			return nil, err
		}

		if inclusive {
			tp = store.RangeROpen
		}
	} else {
		if minKey, err = db.encodeScanMinKey(storeDataType, nil); err != nil {
			return nil, err
		}
		if maxKey, err = db.encodeScanMaxKey(storeDataType, key); err != nil {
			return nil, err
		}

		if inclusive {
			tp = store.RangeLOpen
		}
	}

	if count <= 0 {
		count = defaultScanCount
	}

	var it *store.RangeLimitIterator
	if !reverse {
		it = db.bucket.RangeIterator(minKey, maxKey, tp)
	} else {
		it = db.bucket.RevRangeIterator(minKey, maxKey, tp)
	}

	v := make([][]byte, 0, count)

	for i := 0; it.Valid() && i < count; it.Next() {
		if k, err := db.decodeScanKey(storeDataType, it.Key()); err != nil {
			continue
		} else if r != nil && !r.Match(k) {
			continue
		} else {
			v = append(v, k)
			i++
		}
	}
	it.Close()
	return v, nil
}

func (db *DB) encodeScanMinKey(storeDataType byte, key []byte) ([]byte, error) {
	if len(key) == 0 {
		return db.encodeScanKey(storeDataType, nil)
	} else {
		if err := checkKeySize(key); err != nil {
			return nil, err
		}
		return db.encodeScanKey(storeDataType, key)
	}
}

func (db *DB) encodeScanMaxKey(storeDataType byte, key []byte) ([]byte, error) {
	if len(key) > 0 {
		if err := checkKeySize(key); err != nil {
			return nil, err
		}

		return db.encodeScanKey(storeDataType, key)
	}

	k, err := db.encodeScanKey(storeDataType, nil)
	if err != nil {
		return nil, err
	}
	k[len(k)-1] = storeDataType + 1
	return k, nil
}

func (db *DB) encodeScanKey(storeDataType byte, key []byte) ([]byte, error) {
	switch storeDataType {
	case KVType:
		return db.encodeKVKey(key), nil
	case LMetaType:
		return db.lEncodeMetaKey(key), nil
	case HSizeType:
		return db.hEncodeSizeKey(key), nil
	case ZSizeType:
		return db.zEncodeSizeKey(key), nil
	case SSizeType:
		return db.sEncodeSizeKey(key), nil
	// case BitMetaType:
	// 	return db.bEncodeMetaKey(key), nil
	default:
		return nil, errDataType
	}
}
func (db *DB) decodeScanKey(storeDataType byte, ek []byte) ([]byte, error) {
	if len(ek) < 2 || ek[0] != db.index || ek[1] != storeDataType {
		return nil, errMetaKey
	}
	return ek[2:], nil
}

// for specail data scan

func (db *DB) buildDataScanIterator(start []byte, stop []byte, inclusive bool) *store.RangeLimitIterator {
	tp := store.RangeROpen

	if !inclusive {
		tp = store.RangeOpen
	}
	it := db.bucket.RangeIterator(start, stop, tp)
	return it

}

func (db *DB) HScan(key []byte, cursor []byte, count int, inclusive bool, match string) ([]FVPair, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	start := db.hEncodeHashKey(key, cursor)
	stop := db.hEncodeStopKey(key)

	v := make([]FVPair, 0, 16)

	r, err := buildMatchRegexp(match)
	if err != nil {
		return nil, err
	}

	it := db.buildDataScanIterator(start, stop, inclusive)
	defer it.Close()

	for i := 0; it.Valid() && i < count; it.Next() {
		_, f, err := db.hDecodeHashKey(it.Key())
		if err != nil {
			return nil, err
		} else if r != nil && !r.Match(f) {
			continue
		}

		v = append(v, FVPair{Field: f, Value: it.Value()})

		i++
	}

	return v, nil
}

func (db *DB) SScan(key []byte, cursor []byte, count int, inclusive bool, match string) ([][]byte, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	start := db.sEncodeSetKey(key, cursor)
	stop := db.sEncodeStopKey(key)

	v := make([][]byte, 0, 16)

	r, err := buildMatchRegexp(match)
	if err != nil {
		return nil, err
	}

	it := db.buildDataScanIterator(start, stop, inclusive)
	defer it.Close()

	for i := 0; it.Valid() && i < count; it.Next() {
		_, m, err := db.sDecodeSetKey(it.Key())
		if err != nil {
			return nil, err
		} else if r != nil && !r.Match(m) {
			continue
		}

		v = append(v, m)

		i++
	}

	return v, nil
}

func (db *DB) ZScan(key []byte, cursor []byte, count int, inclusive bool, match string) ([]ScorePair, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	start := db.zEncodeSetKey(key, cursor)
	stop := db.zEncodeStopSetKey(key)

	v := make([]ScorePair, 0, 16)

	r, err := buildMatchRegexp(match)
	if err != nil {
		return nil, err
	}

	it := db.buildDataScanIterator(start, stop, inclusive)
	defer it.Close()

	for i := 0; it.Valid() && i < count; it.Next() {
		_, m, err := db.zDecodeSetKey(it.Key())
		if err != nil {
			return nil, err
		} else if r != nil && !r.Match(m) {
			continue
		}

		score, err := Int64(it.Value(), nil)
		if err != nil {
			return nil, err
		}

		v = append(v, ScorePair{Score: score, Member: m})

		i++
	}

	return v, nil
}
