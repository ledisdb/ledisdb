// +build leveldb

package leveldb

// #cgo LDFLAGS: -lleveldb
// #include "leveldb/c.h"
import "C"

import (
	"github.com/siddontang/goleveldb/leveldb"
	"unsafe"
)

/*
	It's not easy to let leveldb support Data() function like rocksdb,
	so here, we will use goleveldb batch instead

	later optimize
*/

type WriteBatch struct {
	db     *DB
	wbatch *C.leveldb_writebatch_t

	gbatch *leveldb.Batch
}

func newWriteBatch(db *DB) *WriteBatch {
	w := new(WriteBatch)
	w.db = db
	w.wbatch = C.leveldb_writebatch_create()
	w.gbatch = new(leveldb.Batch)

	return w
}

func (w *WriteBatch) Close() {
	if w.wbatch != nil {
		C.leveldb_writebatch_destroy(w.wbatch)
		w.wbatch = nil
	}

	w.gbatch = nil
}

func (w *WriteBatch) Put(key, value []byte) {
	w.gbatch.Put(key, value)

	var k, v *C.char
	if len(key) != 0 {
		k = (*C.char)(unsafe.Pointer(&key[0]))
	}
	if len(value) != 0 {
		v = (*C.char)(unsafe.Pointer(&value[0]))
	}

	lenk := len(key)
	lenv := len(value)

	C.leveldb_writebatch_put(w.wbatch, k, C.size_t(lenk), v, C.size_t(lenv))
}

func (w *WriteBatch) Delete(key []byte) {
	w.gbatch.Delete(key)

	C.leveldb_writebatch_delete(w.wbatch,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)))
}

func (w *WriteBatch) Commit() error {
	w.gbatch.Reset()

	return w.commit(w.db.writeOpts)
}

func (w *WriteBatch) SyncCommit() error {
	w.gbatch.Reset()

	return w.commit(w.db.syncOpts)
}

func (w *WriteBatch) Rollback() error {
	C.leveldb_writebatch_clear(w.wbatch)

	w.gbatch.Reset()

	return nil
}

func (w *WriteBatch) commit(wb *WriteOptions) error {
	var errStr *C.char
	C.leveldb_write(w.db.db, wb.Opt, w.wbatch, &errStr)
	if errStr != nil {
		return saveError(errStr)
	}
	return nil
}

func (w *WriteBatch) Data() []byte {
	return w.gbatch.Dump()
}
