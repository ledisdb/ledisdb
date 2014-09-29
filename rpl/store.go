package rpl

import (
	"errors"
)

const (
	InvalidLogID uint64 = 0
)

var (
	ErrLogNotFound = errors.New("log not found")
	ErrLessLogID   = errors.New("log id is less")
	ErrNoBehindLog = errors.New("no behind commit log")
)

type LogStore interface {
	GetLog(id uint64, log *Log) error

	// Get the first log which ID is equal or larger than id
	SeekLog(id uint64, log *Log) error

	FirstID() (uint64, error)
	LastID() (uint64, error)

	// if log id is less than current last id, return error
	StoreLog(log *Log) error
	StoreLogs(logs []*Log) error

	// Delete first n logs
	Purge(n uint64) error

	// Delete logs before n seconds
	PurgeExpired(n int64) error

	// Clear all logs
	Clear() error

	Close() error
}
