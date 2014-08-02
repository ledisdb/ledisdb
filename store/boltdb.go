// +build !windows

package store

import (
	"github.com/siddontang/copier"
	"github.com/siddontang/ledisdb/store/boltdb"
	"github.com/siddontang/ledisdb/store/driver"
)

const BoltDBName = "boltdb"

type BoltDBStore struct {
}

func (s BoltDBStore) Open(cfg *Config) (driver.IDB, error) {
	c := &boltdb.Config{}
	copier.Copy(c, cfg)

	return boltdb.Open(c)
}

func (s BoltDBStore) Repair(cfg *Config) error {
	c := &boltdb.Config{}
	copier.Copy(c, cfg)

	return boltdb.Repair(c)
}

func init() {
	Register(BoltDBName, BoltDBStore{})
}
