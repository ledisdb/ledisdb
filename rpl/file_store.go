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

func NewFileStore(base string, maxSize int64) (*FileStore, error) {
	s := new(FileStore)

	var err error

	if err = os.MkdirAll(base, 0755); err != nil {
		return nil, err
	}

	s.base = base

	s.maxFileSize = num.MinInt64(maxLogFileSize, maxSize)

	if err = s.load(); err != nil {
		return nil, err
	}

	index := int64(1)
	if len(s.rs) != 0 {
		index = s.rs[len(s.rs)-1].index + 1
	}

	s.w = newTableWriter(s.base, index, s.maxFileSize)
	return s, nil
}

func (s *FileStore) GetLog(id uint64, log *Log) error {
	panic("not implementation")
	return nil
}

func (s *FileStore) FirstID() (uint64, error) {
	return 0, nil
}

func (s *FileStore) LastID() (uint64, error) {
	return 0, nil
}

func (s *FileStore) StoreLog(l *Log) error {
	s.wm.Lock()
	defer s.wm.Unlock()

	if s.w == nil {
		return fmt.Errorf("nil table writer, cannot store")
	}

	err := s.w.StoreLog(l)
	if err == nil {
		return nil
	} else if err != errTableNeedFlush {
		return err
	}

	var r *tableReader
	if r, err = s.w.Flush(); err != nil {
		log.Error("write table flush error %s, can not store now", err.Error())

		s.w.Close()
		s.w = nil
		return err
	}

	s.rm.Lock()
	s.rs = append(s.rs, r)
	s.rm.Unlock()

	return nil
}

func (s *FileStore) PuregeExpired(n int64) error {
	s.rm.Lock()

	purges := []*tableReader{}
	lefts := []*tableReader{}

	t := uint32(time.Now().Unix() - int64(n))

	for _, r := range s.rs {
		if r.lastTime < t {
			purges = append(purges, r)
		} else {
			lefts = append(lefts, r)
		}
	}

	s.rs = lefts

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

func (s *FileStore) Clear() error {
	return nil
}

func (s *FileStore) Close() error {
	s.wm.Lock()
	if s.w != nil {
		if r, err := s.w.Flush(); err != nil {
			log.Error("close err: %s", err.Error())
		} else {
			r.Close()
			s.w.Close()
			s.w = nil
		}
	}

	s.wm.Unlock()

	s.rm.Lock()

	for i := range s.rs {
		s.rs[i].Close()
	}
	s.rs = nil

	s.rm.Unlock()

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
	return ts[i].index < ts[j].index
}

func (ts tableReaders) Search(id uint64) *tableReader {
	n := sort.Search(len(ts), func(i int) bool {
		return id >= ts[i].first && id <= ts[i].last
	})

	if n < len(ts) {
		return ts[n]
	} else {
		return nil
	}
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

		if ts[i].index == index {
			return fmt.Errorf("invalid index %d in table %s", ts[i].index, ts[i].name)
		}

		first = ts[i].first
		last = ts[i].last
		index = ts[i].index
	}
	return nil
}
