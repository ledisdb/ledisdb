// +build leveldb

package store

import (
	"github.com/siddontang/copier"
	"github.com/siddontang/ledisdb/store/driver"
	"github.com/siddontang/ledisdb/store/leveldb"
)

const LevelDBName = "leveldb"

type LevelDBStore struct {
}

func (s LevelDBStore) Open(cfg *Config) (driver.IDB, error) {
	c := &leveldb.Config{}
	copier.Copy(c, cfg)

	return leveldb.Open(c)
}

func (s LevelDBStore) Repair(cfg *Config) error {
	c := &leveldb.Config{}
	copier.Copy(c, cfg)

	return leveldb.Repair(c)
}

func init() {
	Register(LevelDBName, LevelDBStore{})
}
