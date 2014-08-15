package store

import (
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/store/driver"
	"os"
	"path"

	"github.com/siddontang/ledisdb/store/boltdb"
	"github.com/siddontang/ledisdb/store/goleveldb"
	"github.com/siddontang/ledisdb/store/hyperleveldb"
	"github.com/siddontang/ledisdb/store/leveldb"
	"github.com/siddontang/ledisdb/store/mdb"
	"github.com/siddontang/ledisdb/store/rocksdb"
)

func getStorePath(cfg *config.Config) string {
	return path.Join(cfg.DataDir, fmt.Sprintf("%s_data", cfg.DBName))
}

func Open(cfg *config.Config) (*DB, error) {
	s, err := driver.GetStore(cfg)
	if err != nil {
		return nil, err
	}

	path := getStorePath(cfg)

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}

	idb, err := s.Open(path, cfg)
	if err != nil {
		return nil, err
	}

	db := &DB{idb}

	return db, nil
}

func Repair(cfg *config.Config) error {
	s, err := driver.GetStore(cfg)
	if err != nil {
		return err
	}

	path := getStorePath(cfg)

	return s.Repair(path, cfg)
}

func init() {
	_ = boltdb.DBName
	_ = goleveldb.DBName
	_ = hyperleveldb.DBName
	_ = leveldb.DBName
	_ = mdb.DBName
	_ = rocksdb.DBName
}
