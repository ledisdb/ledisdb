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
	eliminator.regRetireContext(kvType, db.kvTx, db.delete)
	eliminator.regRetireContext(listType, db.listTx, db.lDelete)
	eliminator.regRetireContext(hashType, db.hashTx, db.hDelete)
	eliminator.regRetireContext(zsetType, db.zsetTx, db.zDelete)
	eliminator.regRetireContext(binType, db.binTx, db.bDelete)

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
