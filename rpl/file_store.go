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
	defaultMaxLogFileSize = 1024 * 1024 * 1024
)

/*
index file format:
ledis-bin.00001
ledis-bin.00002
ledis-bin.00003
*/

type FileStore struct {
	LogStore

	m sync.Mutex

	maxFileSize int

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

func (s *FileStore) SetMaxFileSize(size int) {
	s.maxFileSize = size
}

func (s *FileStore) GetLog(id uint64, log *Log) error {
	panic("not implementation")
	return nil
}

func (s *FileStore) SeekLog(id uint64, log *Log) error {
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

func (s *FileStore) StoreLogs(logs []*Log) error {
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
