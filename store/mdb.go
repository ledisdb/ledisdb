// +build !windows

package store

import (
	"github.com/siddontang/copier"
	"github.com/siddontang/ledisdb/store/driver"
	"github.com/siddontang/ledisdb/store/mdb"
)

const LMDBName = "lmdb"

type LMDBStore struct {
}

func (s LMDBStore) Open(cfg *Config) (driver.IDB, error) {
	c := &mdb.Config{}
	copier.Copy(c, cfg)

	return mdb.Open(c)
}

func (s LMDBStore) Repair(cfg *Config) error {
	c := &mdb.Config{}
	copier.Copy(c, cfg)

	return mdb.Repair(c)
}

func init() {
	Register(LMDBName, LMDBStore{})
}
