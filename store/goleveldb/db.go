package goleveldb

import (
	"github.com/siddontang/ledisdb/store/driver"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/cache"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
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

type DB struct {
	cfg *Config

	db *leveldb.DB

	opts *opt.Options

	iteratorOpts *opt.ReadOptions

	cache cache.Cache

	filter filter.Filter
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
	db, err := leveldb.RecoverFile(cfg.Path, newOptions(cfg))
	if err != nil {
		return err
	}

	db.Close()
	return nil
}

func (db *DB) open() error {
	db.opts = newOptions(db.cfg)

	db.iteratorOpts = &opt.ReadOptions{}
	db.iteratorOpts.DontFillCache = true

	var err error
	db.db, err = leveldb.OpenFile(db.cfg.Path, db.opts)

	return err
}

func newOptions(cfg *Config) *opt.Options {
	opts := &opt.Options{}
	opts.ErrorIfMissing = false

	if cfg.CacheSize > 0 {
		opts.BlockCache = cache.NewLRUCache(cfg.CacheSize)
	}

	//we must use bloomfilter
	opts.Filter = filter.NewBloomFilter(defaultFilterBits)

	if !cfg.Compression {
		opts.Compression = opt.NoCompression
	} else {
		opts.Compression = opt.SnappyCompression
	}

	if cfg.BlockSize > 0 {
		opts.BlockSize = cfg.BlockSize
	}

	if cfg.WriteBufferSize > 0 {
		opts.WriteBuffer = cfg.WriteBufferSize
	}

	return opts
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Put(key, value []byte) error {
	return db.db.Put(key, value, nil)
}

func (db *DB) Get(key []byte) ([]byte, error) {
	v, err := db.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	return v, nil
}

func (db *DB) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

func (db *DB) NewWriteBatch() driver.IWriteBatch {
	wb := &WriteBatch{
		db:     db,
		wbatch: new(leveldb.Batch),
	}
	return wb
}

func (db *DB) NewIterator() driver.IIterator {
	it := &Iterator{
		db.db.NewIterator(nil, db.iteratorOpts),
	}

	return it
}