package boltdb

import (
	"github.com/boltdb/bolt"
	"github.com/siddontang/ledisdb/store/driver"
	"os"
	"path"
)

var bucketName = []byte("ledisdb")

type Config struct {
	Path   string `json:"path"`
	NoSync bool   `json:"nosync"`
}

type DB struct {
	cfg *Config
	db  *bolt.DB
}

func Open(cfg *Config) (*DB, error) {
	os.MkdirAll(cfg.Path, os.ModePerm)
	name := path.Join(cfg.Path, "ledis_bolt.db")
	db := new(DB)
	var err error
	db.db, err = bolt.Open(name, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.db.NoSync = cfg.NoSync

	var tx *bolt.Tx
	tx, err = db.db.Begin(true)
	if err != nil {
		return nil, err
	}

	_, err = tx.CreateBucketIfNotExists(bucketName)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return db, nil
}

func Repair(cfg *Config) error {
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Get(key []byte) ([]byte, error) {
	var value []byte

	t, err := db.db.Begin(false)
	if err != nil {
		return nil, err
	}
	b := t.Bucket(bucketName)

	value = b.Get(key)
	err = t.Rollback()

	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	} else {
		return append([]byte{}, value...), nil
	}
}

func (db *DB) Put(key []byte, value []byte) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put(key, value)
	})
	return err
}

func (db *DB) Delete(key []byte) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete(key)
	})
	return err

}

func (db *DB) NewIterator() driver.IIterator {
	tx, err := db.db.Begin(false)
	if err != nil {
		return &Iterator{}
	}
	b := tx.Bucket(bucketName)

	return &Iterator{
		tx: tx,
		it: b.Cursor()}
}

func (db *DB) NewWriteBatch() driver.IWriteBatch {
	return driver.NewWriteBatch(db)
}

func (db *DB) Begin() (driver.Tx, error) {
	tx, err := db.db.Begin(true)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx: tx,
		b:  tx.Bucket(bucketName),
	}, nil
}

func (db *DB) BatchPut(writes []driver.Write) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		var err error
		for _, w := range writes {
			if w.Value == nil {
				err = b.Delete(w.Key)
			} else {
				err = b.Put(w.Key, w.Value)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
