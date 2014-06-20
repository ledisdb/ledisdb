package leveldb

// #cgo LDFLAGS: -lleveldb
// #include <stdint.h>
// #include "leveldb/c.h"
import "C"

type Snapshot struct {
	db *DB

	snap *C.leveldb_snapshot_t

	readOpts     *ReadOptions
	iteratorOpts *ReadOptions
}

func (s *Snapshot) Close() {
	C.leveldb_release_snapshot(s.db.db, s.snap)

	s.iteratorOpts.Close()
	s.readOpts.Close()
}

func (s *Snapshot) Get(key []byte) ([]byte, error) {
	return s.db.get(s.readOpts, key)
}

func (s *Snapshot) NewIterator() *Iterator {
	it := new(Iterator)

	it.it = C.leveldb_create_iterator(s.db.db, s.iteratorOpts.Opt)

	return it
}

func (s *Snapshot) RangeIterator(min []byte, max []byte, rangeType uint8) *RangeLimitIterator {
	return newRangeLimitIterator(s.NewIterator(), &Range{min, max, rangeType}, 0, -1, IteratorForward)
}

func (s *Snapshot) RevRangeIterator(min []byte, max []byte, rangeType uint8) *RangeLimitIterator {
	return newRangeLimitIterator(s.NewIterator(), &Range{min, max, rangeType}, 0, -1, IteratorBackward)
}

//limit < 0, unlimit
//offset must >= 0, if < 0, will get nothing
func (s *Snapshot) RangeLimitIterator(min []byte, max []byte, rangeType uint8, offset int, limit int) *RangeLimitIterator {
	return newRangeLimitIterator(s.NewIterator(), &Range{min, max, rangeType}, offset, limit, IteratorForward)
}

//limit < 0, unlimit
//offset must >= 0, if < 0, will get nothing
func (s *Snapshot) RevRangeLimitIterator(min []byte, max []byte, rangeType uint8, offset int, limit int) *RangeLimitIterator {
	return newRangeLimitIterator(s.NewIterator(), &Range{min, max, rangeType}, offset, limit, IteratorBackward)
}
