// +build rocksdb

package rocksdb

// #cgo LDFLAGS: -lrocksdb
// #include "rocksdb/c.h"
import "C"

import (
	"unsafe"
)

type WriteBatch struct {
	db     *DB
	wbatch *C.rocksdb_writebatch_t
}

func (w *WriteBatch) Close() error {
	C.rocksdb_writebatch_destroy(w.wbatch)
	w.wbatch = nil
	return nil
}

func (w *WriteBatch) Put(key, value []byte) {
	var k, v *C.char
	if len(key) != 0 {
		k = (*C.char)(unsafe.Pointer(&key[0]))
	}
	if len(value) != 0 {
		v = (*C.char)(unsafe.Pointer(&value[0]))
	}

	lenk := len(key)
	lenv := len(value)

	C.rocksdb_writebatch_put(w.wbatch, k, C.size_t(lenk), v, C.size_t(lenv))
}

func (w *WriteBatch) Delete(key []byte) {
	C.rocksdb_writebatch_delete(w.wbatch,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)))
}

func (w *WriteBatch) Commit() error {
	return w.commit(w.db.writeOpts)
}

func (w *WriteBatch) SyncCommit() error {
	return w.commit(w.db.syncOpts)
}

func (w *WriteBatch) Rollback() error {
	C.rocksdb_writebatch_clear(w.wbatch)
	return nil
}

func (w *WriteBatch) commit(wb *WriteOptions) error {
	var errStr *C.char
	C.rocksdb_write(w.db.db, wb.Opt, w.wbatch, &errStr)
	if errStr != nil {
		return saveError(errStr)
	}
	return nil
}
