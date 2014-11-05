package rpl

import (
	"fmt"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/ioutil2"
	"github.com/siddontang/go/log"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

const (
	defaultMaxLogFileSize = uint32(1024 * 1024 * 1024)

	//why 4G, we can use uint32 as offset, reduce memory useage
	maxLogFileSize = uint32(4*1024*1024*1024 - 1)

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
	split data = log0 data
	log0: id 0, create time 0, compression 0, data len 0, data ""

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

	m sync.Mutex

	maxFileSize uint32

	first uint64
	last  uint64

	logFile      *os.File
	logNames     []string
	nextLogIndex int64

	indexName string

	path string
}

func NewFileStore(path string) (*FileStore, error) {
	s := new(FileStore)

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	s.path = path

	s.maxFileSize = defaultMaxLogFileSize

	s.first = 0
	s.last = 0

	s.logNames = make([]string, 0, 16)

	if err := s.loadIndex(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *FileStore) SetMaxFileSize(size uint32) {
	s.maxFileSize = size
}

func (s *FileStore) GetLog(id uint64, log *Log) error {
	panic("not implementation")
	return nil
}

func (s *FileStore) FirstID() (uint64, error) {
	panic("not implementation")
	return 0, nil
}

func (s *FileStore) LastID() (uint64, error) {
	panic("not implementation")
	return 0, nil
}

func (s *FileStore) StoreLog(log *Log) error {
	panic("not implementation")
	return nil
}

func (s *FileStore) Purge(n uint64) error {
	panic("not implementation")
	return nil
}

func (s *FileStore) PuregeExpired(n int64) error {
	panic("not implementation")
	return nil
}

func (s *FileStore) Clear() error {
	panic("not implementation")
	return nil
}

func (s *FileStore) Close() error {
	panic("not implementation")
	return nil
}

func (s *FileStore) flushIndex() error {
	data := strings.Join(s.logNames, "\n")

	if err := ioutil2.WriteFileAtomic(s.indexName, hack.Slice(data), 0644); err != nil {
		log.Error("flush index error %s", err.Error())
		return err
	}

	return nil
}

func (s *FileStore) fileExists(name string) bool {
	p := path.Join(s.path, name)
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

func (s *FileStore) loadIndex() error {
	s.indexName = path.Join(s.path, fmt.Sprintf("ledis-bin.index"))
	if _, err := os.Stat(s.indexName); os.IsNotExist(err) {
		//no index file, nothing to do
	} else {
		indexData, err := ioutil.ReadFile(s.indexName)
		if err != nil {
			return err
		}

		lines := strings.Split(string(indexData), "\n")
		for _, line := range lines {
			line = strings.Trim(line, "\r\n ")
			if len(line) == 0 {
				continue
			}

			if s.fileExists(line) {
				s.logNames = append(s.logNames, line)
			} else {
				log.Info("log %s has not exists", line)
			}
		}
	}

	var err error
	if len(s.logNames) == 0 {
		s.nextLogIndex = 1
	} else {
		lastName := s.logNames[len(s.logNames)-1]

		if s.nextLogIndex, err = strconv.ParseInt(path.Ext(lastName)[1:], 10, 64); err != nil {
			log.Error("invalid logfile name %s", err.Error())
			return err
		}

		//like mysql, if server restart, a new log will create
		s.nextLogIndex++
	}

	return nil
}

func (s *FileStore) openNewLogFile() error {
	var err error
	lastName := s.formatLogFileName(s.nextLogIndex)

	logPath := path.Join(s.path, lastName)
	if s.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Error("open new logfile error %s", err.Error())
		return err
	}

	s.logNames = append(s.logNames, lastName)

	if err = s.flushIndex(); err != nil {
		return err
	}

	return nil
}

func (s *FileStore) checkLogFileSize() bool {
	if s.logFile == nil {
		return false
	}

	st, _ := s.logFile.Stat()
	if st.Size() >= int64(s.maxFileSize) {
		s.closeLog()
		return true
	}

	return false
}

func (s *FileStore) closeLog() {
	if s.logFile == nil {
		return
	}

	s.nextLogIndex++

	s.logFile.Close()
	s.logFile = nil
}

func (s *FileStore) formatLogFileName(index int64) string {
	return fmt.Sprintf("ledis-bin.%07d", index)
}
