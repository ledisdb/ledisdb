package wal

import (
	"bytes"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/store"
	"os"
	"sync"
)

type GoLevelDBStore struct {
	m  sync.Mutex
	db *store.DB

	cfg *config.Config

	first uint64
	last  uint64
}

func (s *GoLevelDBStore) FirstID() (uint64, error) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.firstID()
}

func (s *GoLevelDBStore) LastID() (uint64, error) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.lastID()
}

func (s *GoLevelDBStore) firstID() (uint64, error) {
	if s.first != InvalidLogID {
		return s.first, nil
	}

	it := s.db.NewIterator()
	defer it.Close()

	it.SeekToFirst()

	if it.Valid() {
		s.first = num.BytesToUint64(it.RawKey())
	}

	return s.first, nil
}

func (s *GoLevelDBStore) lastID() (uint64, error) {
	if s.last != InvalidLogID {
		return s.last, nil
	}

	it := s.db.NewIterator()
	defer it.Close()

	it.SeekToLast()

	if it.Valid() {
		s.last = num.BytesToUint64(it.RawKey())
	}

	return s.last, nil
}

func (s *GoLevelDBStore) GetLog(id uint64, log *Log) error {
	v, err := s.db.Get(num.Uint64ToBytes(id))
	if err != nil {
		return err
	} else if v == nil {
		return ErrLogNotFound
	} else {
		return log.Decode(bytes.NewBuffer(v))
	}
}

func (s *GoLevelDBStore) SeekLog(id uint64, log *Log) error {
	it := s.db.NewIterator()
	defer it.Close()

	it.Seek(num.Uint64ToBytes(id))

	if !it.Valid() {
		return ErrLogNotFound
	} else {
		return log.Decode(bytes.NewBuffer(it.RawValue()))
	}
}

func (s *GoLevelDBStore) StoreLog(log *Log) error {
	return s.StoreLogs([]*Log{log})
}

func (s *GoLevelDBStore) StoreLogs(logs []*Log) error {
	s.m.Lock()
	defer s.m.Unlock()

	w := s.db.NewWriteBatch()
	defer w.Rollback()

	last, err := s.lastID()
	if err != nil {
		return err
	}

	s.last = InvalidLogID

	var buf bytes.Buffer
	for _, log := range logs {
		buf.Reset()

		if log.ID <= last {
			return ErrLessLogID
		}

		last = log.ID
		key := num.Uint64ToBytes(log.ID)

		if err := log.Encode(&buf); err != nil {
			return err
		}
		w.Put(key, buf.Bytes())
	}

	if err := w.Commit(); err != nil {
		return err
	}

	s.last = last
	return nil
}

func (s *GoLevelDBStore) DeleteRange(min, max uint64) error {
	s.m.Lock()
	defer s.m.Unlock()

	var first, last uint64
	var err error

	first, err = s.firstID()
	if err != nil {
		return err
	}

	last, err = s.lastID()
	if err != nil {
		return err
	}

	min = num.MaxUint64(min, first)
	max = num.MinUint64(max, last)

	w := s.db.NewWriteBatch()
	defer w.Rollback()

	n := 0

	s.reset()

	for i := min; i <= max; i++ {
		w.Delete(num.Uint64ToBytes(i))
		n++
		if n > 1024 {
			if err = w.Commit(); err != nil {
				return err
			}
			n = 0
		}
	}

	if err = w.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *GoLevelDBStore) Clear() error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.db != nil {
		s.db.Close()
	}

	s.reset()
	os.RemoveAll(s.cfg.DBPath)

	return s.open()
}

func (s *GoLevelDBStore) reset() {
	s.first = InvalidLogID
	s.last = InvalidLogID
}

func (s *GoLevelDBStore) Close() error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.db == nil {
		return nil
	}

	err := s.db.Close()
	s.db = nil
	return err
}

func (s *GoLevelDBStore) open() error {
	var err error

	s.first = InvalidLogID
	s.last = InvalidLogID

	s.db, err = store.Open(s.cfg)
	return err
}

func NewGoLevelDBStore(base string) (*GoLevelDBStore, error) {
	cfg := new(config.Config)
	cfg.DBName = "goleveldb"
	cfg.DBPath = base
	cfg.LevelDB.BlockSize = 4 * 1024 * 1024
	cfg.LevelDB.CacheSize = 16 * 1024 * 1024
	cfg.LevelDB.WriteBufferSize = 4 * 1024 * 1024
	cfg.LevelDB.Compression = false

	s := new(GoLevelDBStore)
	s.cfg = cfg

	if err := s.open(); err != nil {
		return nil, err
	}

	return s, nil
}
