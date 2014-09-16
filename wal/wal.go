package wal

import (
	"errors"
)

const (
	InvalidLogID uint64 = 0
)

var (
	ErrLogNotFound = errors.New("log not found")
	ErrLessLogID   = errors.New("log id is less")
)

type Store interface {
	GetLog(id uint64, log *Log) error

	// Get the first log which ID is equal or larger than id
	SeekLog(id uint64, log *Log) error

	FirstID() (uint64, error)
	LastID() (uint64, error)

	// if log id is less than current last id, return error
	StoreLog(log *Log) error
	StoreLogs(logs []*Log) error

	// Delete logs [start, stop]
	DeleteRange(start, stop uint64) error

	// Clear all logs
	Clear() error

	Close() error
}
