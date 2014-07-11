package ledis

import (
	"github.com/siddontang/ledisdb/leveldb"
)

func (db *DB) FlushAll() (drop int64, err error) {
	all := [...](func() (int64, error)){
		db.flush,
		db.lFlush,
		db.hFlush,
		db.zFlush,
		db.bFlush}

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
	it := db.db.RangeIterator(minKey, maxKey, leveldb.RangeROpen)
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
