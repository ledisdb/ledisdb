package store

import (
	"github.com/siddontang/ledisdb/store/driver"
)

type DB struct {
	db driver.IDB
}

// Close database
//
// Caveat
//  Any other DB operations like Get, Put, etc... may cause a panic after Close
//
func (db *DB) Close() error {
	if db.db == nil {
		return nil
	}

	err := db.db.Close()
	db.db = nil

	return err
}

// Get Value with Key
func (db *DB) Get(key []byte) ([]byte, error) {
	return db.db.Get(key)
}

// Put value with key
func (db *DB) Put(key []byte, value []byte) error {
	err := db.db.Put(key, value)
	return err
}

// Delete by key
func (db *DB) Delete(key []byte) error {
	err := db.db.Delete(key)
	return err
}

func (db *DB) NewIterator() *Iterator {
	it := new(Iterator)
	it.it = db.db.NewIterator()

	return it
}

func (db *DB) NewWriteBatch() *WriteBatch {
	return &WriteBatch{db.db.NewWriteBatch()}
}

func (db *DB) RangeIterator(min []byte, max []byte, rangeType uint8) *RangeLimitIterator {
	return NewRangeLimitIterator(db.NewIterator(), &Range{min, max, rangeType}, &Limit{0, -1})
}

func (db *DB) RevRangeIterator(min []byte, max []byte, rangeType uint8) *RangeLimitIterator {
	return NewRevRangeLimitIterator(db.NewIterator(), &Range{min, max, rangeType}, &Limit{0, -1})
}

//count < 0, unlimit.
//
//offset must >= 0, if < 0, will get nothing.
func (db *DB) RangeLimitIterator(min []byte, max []byte, rangeType uint8, offset int, count int) *RangeLimitIterator {
	return NewRangeLimitIterator(db.NewIterator(), &Range{min, max, rangeType}, &Limit{offset, count})
}

//count < 0, unlimit.
//
//offset must >= 0, if < 0, will get nothing.
func (db *DB) RevRangeLimitIterator(min []byte, max []byte, rangeType uint8, offset int, count int) *RangeLimitIterator {
	return NewRevRangeLimitIterator(db.NewIterator(), &Range{min, max, rangeType}, &Limit{offset, count})
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{tx}, nil
}
