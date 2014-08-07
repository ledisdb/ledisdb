package store

import (
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/store/driver"
	"os"
	"path"
)

type Config config.Config

type Store interface {
	String() string
	Open(path string, cfg *config.Config) (driver.IDB, error)
	Repair(paht string, cfg *config.Config) error
}

var dbs = map[string]Store{}

func Register(s Store) {
	name := s.String()
	if _, ok := dbs[name]; ok {
		panic(fmt.Errorf("store %s is registered", s))
	}

	dbs[name] = s
}

func ListStores() []string {
	s := []string{}
	for k, _ := range dbs {
		s = append(s, k)
	}

	return s
}

func getStore(cfg *config.Config) (Store, error) {
	if len(cfg.DBName) == 0 {
		cfg.DBName = config.DefaultDBName
	}

	s, ok := dbs[cfg.DBName]
	if !ok {
		return nil, fmt.Errorf("store %s is not registered", cfg.DBName)
	}

	return s, nil
}

func getStorePath(cfg *config.Config) string {
	return path.Join(cfg.DataDir, fmt.Sprintf("%s_data", cfg.DBName))
}

func Open(cfg *config.Config) (*DB, error) {
	s, err := getStore(cfg)
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
	s, err := getStore(cfg)
	if err != nil {
		return err
	}

	path := getStorePath(cfg)

	return s.Repair(path, cfg)
}
