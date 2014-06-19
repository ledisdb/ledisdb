package leveldb

// #cgo LDFLAGS: -lleveldb
// #include "leveldb/c.h"
import "C"

import (
	"unsafe"
)

type WriteBatch struct {
	db     *DB
	wbatch *C.leveldb_writebatch_t
}

func (w *WriteBatch) Close() {
	C.leveldb_writebatch_destroy(w.wbatch)
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

	C.leveldb_writebatch_put(w.wbatch, k, C.size_t(lenk), v, C.size_t(lenv))
}

func (w *WriteBatch) Delete(key []byte) {
	C.leveldb_writebatch_delete(w.wbatch,
		(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)))
}

func (w *WriteBatch) Commit() error {
	return w.commit(w.db.writeOpts)
}

func (w *WriteBatch) SyncCommit() error {
	return w.commit(w.db.syncWriteOpts)
}

func (w *WriteBatch) Rollback() {
	C.leveldb_writebatch_clear(w.wbatch)
}

func (w *WriteBatch) commit(wb *WriteOptions) error {
	var errStr *C.char
	C.leveldb_write(w.db.db, wb.Opt, w.wbatch, &errStr)
	if errStr != nil {
		return saveError(errStr)
	}
	return nil
}
