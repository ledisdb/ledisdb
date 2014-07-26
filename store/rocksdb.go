// +build rocksdb

package store

import (
	"github.com/siddontang/copier"
	"github.com/siddontang/ledisdb/store/driver"
	"github.com/siddontang/ledisdb/store/rocksdb"
)

const RocksDBName = "rocksdb"

type RocksDBStore struct {
}

func (s RocksDBStore) Open(cfg *Config) (driver.IDB, error) {
	c := &rocksdb.Config{}
	copier.Copy(c, cfg)

	return rocksdb.Open(c)
}

func (s RocksDBStore) Repair(cfg *Config) error {
	c := &rocksdb.Config{}
	copier.Copy(c, cfg)

	return rocksdb.Repair(c)
}

func init() {
	Register(RocksDBName, RocksDBStore{})
}
