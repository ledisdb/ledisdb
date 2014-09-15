package wal

import (
	"os"
	"sync"
)

const (
	defaultMaxLogFileSize = 1024 * 1024 * 1024
	defaultMaxLogFileNum  = 10
)

type FileStore struct {
	Store

	m sync.Mutex

	maxFileSize int
	maxFileNum  int

	first uint64
	last  uint64
}

func NewFileStore(path string) (*FileStore, error) {
	s := new(FileStore)

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	s.maxFileSize = defaultMaxLogFileSize
	s.maxFileNum = defaultMaxLogFileNum

	s.first = 0
	s.last = 0

	return s, nil
}

func (s *FileStore) SetMaxFileSize(size int) {
	s.maxFileSize = size
}

func (s *FileStore) SetMaxFileNum(n int) {
	s.maxFileNum = n
}

func (s *FileStore) GetLog(id uint64, log *Log) error {
	return nil
}

func (s *FileStore) SeekLog(id uint64, log *Log) error {
	return nil
}

func (s *FileStore) FirstID() (uint64, error) {
	return 0, nil
}

func (s *FileStore) LastID() (uint64, error) {
	return 0, nil
}

func (s *FileStore) StoreLog(log *Log) error {
	return nil
}

func (s *FileStore) StoreLogs(logs []*Log) error {
	return nil
}

func (s *FileStore) DeleteRange(start, stop uint64) error {
	return nil
}

func (s *FileStore) Clear() error {
	return nil
}

func (s *FileStore) Close() error {
	return nil
}
