// +build leveldb

// Package leveldb is a wrapper for c++ leveldb
package leveldb

/*
#cgo LDFLAGS: -lleveldb
#include <leveldb/c.h>
#include "leveldb_ext.h"
*/
import "C"

import (
	"github.com/siddontang/ledisdb/store/driver"
	"os"
	"runtime"
	"unsafe"
)

const defaultFilterBits int = 10

type Config struct {
	Path string `json:"path"`

	Compression     bool `json:"compression"`
	BlockSize       int  `json:"block_size"`
	WriteBufferSize int  `json:"write_buffer_size"`
	CacheSize       int  `json:"cache_size"`
	MaxOpenFiles    int  `json:"max_open_files"`
}

func Open(cfg *Config) (*DB, error) {
	if err := os.MkdirAll(cfg.Path, os.ModePerm); err != nil {
		return nil, err
	}

	db := new(DB)
	db.cfg = cfg

	if err := db.open(); err != nil {
		return nil, err
	}

	return db, nil
}

func Repair(cfg *Config) error {
	db := new(DB)
	db.cfg = cfg

	err := db.open()
	defer db.Close()

	//open ok, do not need repair
	if err == nil {
		return nil
	}

	var errStr *C.char
	ldbname := C.CString(db.cfg.Path)
	defer C.leveldb_free(unsafe.Pointer(ldbname))

	C.leveldb_repair_db(db.opts.Opt, ldbname, &errStr)
	if errStr != nil {
		return saveError(errStr)
	}
	return nil
}

type DB struct {
	cfg *Config

	db *C.leveldb_t

	opts *Options

	//for default read and write options
	readOpts     *ReadOptions
	writeOpts    *WriteOptions
	iteratorOpts *ReadOptions

	cache *Cache

	filter *FilterPolicy
}

func (db *DB) open() error {
	db.initOptions(db.cfg)

	var errStr *C.char
	ldbname := C.CString(db.cfg.Path)
	defer C.leveldb_free(unsafe.Pointer(ldbname))

	db.db = C.leveldb_open(db.opts.Opt, ldbname, &errStr)
	if errStr != nil {
		db.db = nil
		return saveError(errStr)
	}
	return nil
}

func (db *DB) initOptions(cfg *Config) {
	opts := NewOptions()

	opts.SetCreateIfMissing(true)

	if cfg.CacheSize <= 0 {
		cfg.CacheSize = 4 * 1024 * 1024
	}

	db.cache = NewLRUCache(cfg.CacheSize)
	opts.SetCache(db.cache)

	//we must use bloomfilter
	db.filter = NewBloomFilter(defaultFilterBits)
	opts.SetFilterPolicy(db.filter)

	if !cfg.Compression {
		opts.SetCompression(NoCompression)
	} else {
		opts.SetCompression(SnappyCompression)
	}

	if cfg.BlockSize <= 0 {
		cfg.BlockSize = 4 * 1024
	}

	opts.SetBlockSize(cfg.BlockSize)

	if cfg.WriteBufferSize <= 0 {
		cfg.WriteBufferSize = 4 * 1024 * 1024
	}

	opts.SetWriteBufferSize(cfg.WriteBufferSize)

	if cfg.MaxOpenFiles < 1024 {
		cfg.MaxOpenFiles = 1024
	}

	opts.SetMaxOpenFiles(cfg.MaxOpenFiles)

	db.opts = opts

	db.readOpts = NewReadOptions()
	db.writeOpts = NewWriteOptions()

	db.iteratorOpts = NewReadOptions()
	db.iteratorOpts.SetFillCache(false)
}

func (db *DB) Close() error {
	if db.db != nil {
		C.leveldb_close(db.db)
		db.db = nil
	}

	db.opts.Close()

	if db.cache != nil {
		db.cache.Close()
	}

	if db.filter != nil {
		db.filter.Close()
	}

	db.readOpts.Close()
	db.writeOpts.Close()
	db.iteratorOpts.Close()

	return nil
}

func (db *DB) Put(key, value []byte) error {
	return db.put(db.writeOpts, key, value)
}

func (db *DB) Get(key []byte) ([]byte, error) {
	return db.get(db.readOpts, key)
}

func (db *DB) Delete(key []byte) error {
	return db.delete(db.writeOpts, key)
}

func (db *DB) NewWriteBatch() driver.IWriteBatch {
	wb := &WriteBatch{
		db:     db,
		wbatch: C.leveldb_writebatch_create(),
	}

	runtime.SetFinalizer(wb, func(w *WriteBatch) {
		w.Close()
	})
	return wb
}

func (db *DB) NewIterator() driver.IIterator {
	it := new(Iterator)

	it.it = C.leveldb_create_iterator(db.db, db.iteratorOpts.Opt)

	return it
}

func (db *DB) put(wo *WriteOptions, key, value []byte) error {
	var errStr *C.char
	var k, v *C.char
	if len(key) != 0 {
		k = (*C.char)(unsafe.Pointer(&key[0]))
	}
	if len(value) != 0 {
		v = (*C.char)(unsafe.Pointer(&value[0]))
	}

	lenk := len(key)
	lenv := len(value)
	C.leveldb_put(
		db.db, wo.Opt, k, C.size_t(lenk), v, C.size_t(lenv), &errStr)

	if errStr != nil {
		return saveError(errStr)
	}
	return nil
}

func (db *DB) get(ro *ReadOptions, key []byte) ([]byte, error) {
	var errStr *C.char
	var vallen C.size_t
	var k *C.char
	if len(key) != 0 {
		k = (*C.char)(unsafe.Pointer(&key[0]))
	}

	var value *C.char

	c := C.leveldb_get_ext(
		db.db, ro.Opt, k, C.size_t(len(key)), &value, &vallen, &errStr)

	if errStr != nil {
		return nil, saveError(errStr)
	}

	if value == nil {
		return nil, nil
	}

	defer C.leveldb_get_free_ext(unsafe.Pointer(c))

	return C.GoBytes(unsafe.Pointer(value), C.int(vallen)), nil
}

func (db *DB) delete(wo *WriteOptions, key []byte) error {
	var errStr *C.char
	var k *C.char
	if len(key) != 0 {
		k = (*C.char)(unsafe.Pointer(&key[0]))
	}

	C.leveldb_delete(
		db.db, wo.Opt, k, C.size_t(len(key)), &errStr)

	if errStr != nil {
		return saveError(errStr)
	}
	return nil
}

func (db *DB) Begin() (driver.Tx, error) {
	return nil, driver.ErrTxSupport
}
