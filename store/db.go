package store

import (
	"github.com/siddontang/ledisdb/store/driver"
	"time"
)

type DB struct {
	driver.IDB
	name string

	st *Stat
}

func (db *DB) String() string {
	return db.name
}

func (db *DB) NewIterator() *Iterator {
	db.st.IterNum.Add(1)

	it := new(Iterator)
	it.it = db.IDB.NewIterator()
	it.st = db.st

	return it
}

func (db *DB) Get(key []byte) ([]byte, error) {
	v, err := db.IDB.Get(key)
	db.st.statGet(v, err)
	return v, err
}

func (db *DB) Put(key []byte, value []byte) error {
	db.st.PutNum.Add(1)
	return db.IDB.Put(key, value)
}

func (db *DB) Delete(key []byte) error {
	db.st.DeleteNum.Add(1)
	return db.IDB.Delete(key)
}

func (db *DB) SyncPut(key []byte, value []byte) error {
	db.st.SyncPutNum.Add(1)
	return db.IDB.SyncPut(key, value)
}

func (db *DB) SyncDelete(key []byte) error {
	db.st.SyncDeleteNum.Add(1)
	return db.IDB.SyncDelete(key)
}

func (db *DB) NewWriteBatch() *WriteBatch {
	db.st.BatchNum.Add(1)
	wb := new(WriteBatch)
	wb.IWriteBatch = db.IDB.NewWriteBatch()
	wb.st = db.st
	return wb
}

func (db *DB) NewSnapshot() (*Snapshot, error) {
	db.st.SnapshotNum.Add(1)

	var err error
	s := &Snapshot{}
	if s.ISnapshot, err = db.IDB.NewSnapshot(); err != nil {
		return nil, err
	}
	s.st = db.st

	return s, nil
}

func (db *DB) Compact() error {
	db.st.CompactNum.Add(1)

	t := time.Now()
	err := db.IDB.Compact()

	db.st.CompactTotalTime.Add(time.Now().Sub(t))

	return err
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
	tx, err := db.IDB.Begin()
	if err != nil {
		return nil, err
	}

	db.st.TxNum.Add(1)

	return &Tx{tx, db.st}, nil
}

func (db *DB) Stat() *Stat {
	return db.st
}
