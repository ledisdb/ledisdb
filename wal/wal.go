package wal

import (
	"errors"
	"github.com/siddontang/ledisdb/config"
	"path"
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

	// Delete first n logs
	Purge(n uint64) error

	// Delete logs before n seconds
	PurgeExpired(n int) error

	// Clear all logs
	Clear() error

	Close() error
}

func NewStore(cfg *config.Config) (Store, error) {
	//now we only support goleveldb

	base := cfg.WAL.Path
	if len(base) == 0 {
		base = path.Join(cfg.DataDir, "wal")
	}

	return NewGoLevelDBStore(base)
}
