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

type LogIDGenerator interface {
	// Force reset to id, if current id is larger than id, nothing reset
	Reset(id uint64) error

	// ID must be first at 1, and increased monotonously, 0 is invalid
	GenerateID() (uint64, error)

	Close() error
}

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
