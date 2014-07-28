package store

import (
	"github.com/siddontang/copier"
	"github.com/siddontang/ledisdb/store/driver"
	"github.com/siddontang/ledisdb/store/goleveldb"
)

const GoLevelDBName = "goleveldb"

type GoLevelDBStore struct {
}

func (s GoLevelDBStore) Open(cfg *Config) (driver.IDB, error) {
	c := &goleveldb.Config{}
	copier.Copy(c, cfg)

	return goleveldb.Open(c)
}

func (s GoLevelDBStore) Repair(cfg *Config) error {
	c := &goleveldb.Config{}
	copier.Copy(c, cfg)

	return goleveldb.Repair(c)
}

func init() {
	Register(GoLevelDBName, GoLevelDBStore{})
}
