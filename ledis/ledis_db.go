package ledis

import (
	"fmt"
	"github.com/siddontang/ledisdb/store"
)

func (db *DB) FlushAll() (drop int64, err error) {
	all := [...](func() (int64, error)){
		db.flush,
		db.lFlush,
		db.hFlush,
		db.zFlush,
		db.bFlush,
		db.sFlush}

	for _, flush := range all {
		if n, e := flush(); e != nil {
			err = e
			return
		} else {
			drop += n
		}
	}

	return
}

func (db *DB) newEliminator() *elimination {
	eliminator := newEliminator(db)
	eliminator.regRetireContext(KVType, db.kvTx, db.delete)
	eliminator.regRetireContext(ListType, db.listTx, db.lDelete)
	eliminator.regRetireContext(HashType, db.hashTx, db.hDelete)
	eliminator.regRetireContext(ZSetType, db.zsetTx, db.zDelete)
	eliminator.regRetireContext(BitType, db.binTx, db.bDelete)

	return eliminator
}

func (db *DB) flushRegion(t *tx, minKey []byte, maxKey []byte) (drop int64, err error) {
	it := db.db.RangeIterator(minKey, maxKey, store.RangeROpen)
	for ; it.Valid(); it.Next() {
		t.Delete(it.RawKey())
		drop++
		if drop&1023 == 0 {
			if err = t.Commit(); err != nil {
				return
			}
		}
	}
	it.Close()
	return
}

func (db *DB) flushType(t *tx, dataType byte) (drop int64, err error) {
	var deleteFunc func(t *tx, key []byte) int64
	var metaDataType byte
	switch dataType {
	case KVType:
		deleteFunc = db.delete
		metaDataType = KVType
	case ListType:
		deleteFunc = db.lDelete
		metaDataType = LMetaType
	case HashType:
		deleteFunc = db.hDelete
		metaDataType = HSizeType
	case ZSetType:
		deleteFunc = db.zDelete
		metaDataType = ZSizeType
	case BitType:
		deleteFunc = db.bDelete
		metaDataType = BitMetaType
	default:
		return 0, fmt.Errorf("invalid data type: %s", TypeName[dataType])
	}

	var keys [][]byte
	keys, err = db.scan(metaDataType, nil, 1024, false)
	for len(keys) != 0 || err != nil {
		for _, key := range keys {
			deleteFunc(t, key)
			db.rmExpire(t, dataType, key)

		}

		if err = t.Commit(); err != nil {
			return
		} else {
			drop += int64(len(keys))
		}
		keys, err = db.scan(metaDataType, nil, 1024, false)
	}
	return
}
