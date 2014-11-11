package rpl

import (
	"fmt"
	"github.com/siddontang/go/log"
	"github.com/siddontang/go/num"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	defaultMaxLogFileSize = int64(1024 * 1024 * 1024)

	//why 4G, we can use uint32 as offset, reduce memory useage
	maxLogFileSize = int64(uint32(4*1024*1024*1024 - 1))

	maxLogNumInFile = uint64(10000000)
)

/*
	File Store:
	00000001.ldb
	00000002.ldb

	log: log1 data | log2 data | split data | log1 offset | log 2 offset | offset start pos | offset length | magic data

	log id can not be 0, we use here for split data
	if data has no magic data, it means that we don't close replication gracefully.
	so we must repair the log data
	log data: id (bigendian uint64), create time (bigendian uint32), compression (byte), data len(bigendian uint32), data
	split data = log0 data + [padding 0] -> file % pagesize() == 0
	log0: id 0, create time 0, compression 0, data len 7, data "ledisdb"

	log offset: bigendian uint32 | bigendian uint32

	offset start pos: bigendian uint64
	offset length: bigendian uint32

	//sha1 of github.com/siddontang/ledisdb 20 bytes
	magic data = "\x1c\x1d\xb8\x88\xff\x9e\x45\x55\x40\xf0\x4c\xda\xe0\xce\x47\xde\x65\x48\x71\x17"

	we must guarantee that the log id is monotonic increment strictly.
	if log1's id is 1, log2 must be 2
*/

type FileStore struct {
	LogStore

	maxFileSize int64

	base string

	rm sync.RWMutex
	wm sync.Mutex

	rs tableReaders
	w  *tableWriter
}

func NewFileStore(base string, maxSize int64, syncType int) (*FileStore, error) {
	s := new(FileStore)

	var err error

	if err = os.MkdirAll(base, 0755); err != nil {
		return nil, err
	}

	s.base = base

	s.maxFileSize = num.MinInt64(maxLogFileSize, maxSize)
	if s.maxFileSize == 0 {
		s.maxFileSize = defaultMaxLogFileSize
	}

	if err = s.load(); err != nil {
		return nil, err
	}

	index := int64(1)
	if len(s.rs) != 0 {
		index = s.rs[len(s.rs)-1].index + 1
	}

	s.w = newTableWriter(s.base, index, s.maxFileSize)
	s.w.SetSyncType(syncType)

	return s, nil
}

func (s *FileStore) GetLog(id uint64, l *Log) error {
	//first search in table writer
	if err := s.w.GetLog(id, l); err == nil {
		return nil
	} else if err != ErrLogNotFound {
		return err
	}

	s.rm.RLock()
	t := s.rs.Search(id)

	if t == nil {
		s.rm.RUnlock()

		return ErrLogNotFound
	}

	err := t.GetLog(id, l)
	s.rm.RUnlock()

	return err
}

func (s *FileStore) FirstID() (uint64, error) {
	id := uint64(0)

	s.rm.RLock()
	if len(s.rs) > 0 {
		id = s.rs[0].first
	} else {
		id = 0
	}
	s.rm.RUnlock()

	if id > 0 {
		return id, nil
	}

	//if id = 0,

	return s.w.First(), nil
}

func (s *FileStore) LastID() (uint64, error) {
	id := s.w.Last()
	if id > 0 {
		return id, nil
	}

	//if table writer has no last id, we may find in the last table reader

	s.rm.RLock()
	if len(s.rs) > 0 {
		id = s.rs[len(s.rs)-1].last
	}
	s.rm.RUnlock()

	return id, nil
}

func (s *FileStore) StoreLog(l *Log) error {
	s.wm.Lock()
	defer s.wm.Unlock()

	err := s.w.StoreLog(l)
	if err == nil {
		return nil
	} else if err != errTableNeedFlush {
		return err
	}

	s.rm.Lock()

	var r *tableReader
	if r, err = s.w.Flush(); err != nil {
		log.Error("write table flush error %s, can not store now", err.Error())

		s.w.Close()

		s.rm.Unlock()

		return err
	}

	s.rs = append(s.rs, r)
	s.rm.Unlock()

	return s.w.StoreLog(l)
}

func (s *FileStore) PuregeExpired(n int64) error {
	s.rm.Lock()

	purges := []*tableReader{}

	t := uint32(time.Now().Unix() - int64(n))

	for i, r := range s.rs {
		if r.lastTime > t {
			purges = s.rs[0:i]
			s.rs = s.rs[i:]
			break
		}
	}

	s.rm.Unlock()

	for _, r := range purges {
		name := r.name
		r.Close()
		if err := os.Remove(name); err != nil {
			log.Error("purge table %s err: %s", name, err.Error())
		}
	}

	return nil
}

func (s *FileStore) Sync() error {
	return s.w.Sync()
}

func (s *FileStore) Clear() error {
	s.wm.Lock()
	s.rm.Lock()

	defer func() {
		s.rm.Unlock()
		s.wm.Unlock()
	}()

	s.w.Close()

	for i := range s.rs {
		s.rs[i].Close()
	}

	s.rs = tableReaders{}

	if err := os.RemoveAll(s.base); err != nil {
		return err
	}

	if err := os.MkdirAll(s.base, 0755); err != nil {
		return err
	}

	s.w = newTableWriter(s.base, 1, s.maxFileSize)

	return nil
}

func (s *FileStore) Close() error {
	s.wm.Lock()
	s.rm.Lock()

	if r, err := s.w.Flush(); err != nil {
		if err != errNilHandler {
			log.Error("close err: %s", err.Error())
		}
	} else {
		r.Close()
		s.w.Close()
	}

	for i := range s.rs {
		s.rs[i].Close()
	}
	s.rs = tableReaders{}

	s.rm.Unlock()
	s.wm.Unlock()

	return nil
}

func (s *FileStore) load() error {
	fs, err := ioutil.ReadDir(s.base)
	if err != nil {
		return err
	}

	var r *tableReader
	var index int64
	for _, f := range fs {
		if _, err := fmt.Sscanf(f.Name(), "%08d.ldb", &index); err == nil {
			if r, err = newTableReader(s.base, index); err != nil {
				log.Error("load table %s err: %s", f.Name(), err.Error())
			} else {
				s.rs = append(s.rs, r)
			}
		}
	}

	if err := s.rs.check(); err != nil {
		return err
	}

	return nil
}

type tableReaders []*tableReader

func (ts tableReaders) Len() int {
	return len(ts)
}

func (ts tableReaders) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (ts tableReaders) Less(i, j int) bool {
	return ts[i].first < ts[j].first
}

func (ts tableReaders) Search(id uint64) *tableReader {
	i, j := 0, len(ts)-1

	for i <= j {
		h := i + (j-i)/2

		if ts[h].first <= id && id <= ts[h].last {
			return ts[h]
		} else if ts[h].last < id {
			i = h + 1
		} else {
			j = h - 1
		}
	}

	return nil
}

func (ts tableReaders) check() error {
	if len(ts) == 0 {
		return nil
	}

	sort.Sort(ts)

	first := ts[0].first
	last := ts[0].last
	index := ts[0].index

	if first == 0 || first > last {
		return fmt.Errorf("invalid log in table %s", ts[0].name)
	}

	for i := 1; i < len(ts); i++ {
		if ts[i].first <= last {
			return fmt.Errorf("invalid first log id %d in table %s", ts[i].first, ts[i].name)
		}

		if ts[i].index <= index {
			return fmt.Errorf("invalid index %d in table %s", ts[i].index, ts[i].name)
		}

		first = ts[i].first
		last = ts[i].last
		index = ts[i].index
	}
	return nil
}
