package ledis

import (
	"encoding/json"
	"github.com/siddontang/go-leveldb/leveldb"
)

type DB struct {
	db *leveldb.DB

	kvTx   *tx
	listTx *tx
	hashTx *tx
	zsetTx *tx
}

func OpenDB(configJson json.RawMessage) (*DB, error) {
	db, err := leveldb.Open(configJson)
	if err != nil {
		return nil, err
	}

	return newDB(db)
}

func OpenDBWithConfig(cfg *leveldb.Config) (*DB, error) {
	db, err := leveldb.OpenWithConfig(cfg)
	if err != nil {
		return nil, err
	}

	return newDB(db)
}

func newDB(db *leveldb.DB) (*DB, error) {
	d := new(DB)

	d.db = db

	d.kvTx = &tx{wb: db.NewWriteBatch()}
	d.listTx = &tx{wb: db.NewWriteBatch()}
	d.hashTx = &tx{wb: db.NewWriteBatch()}
	d.zsetTx = &tx{wb: db.NewWriteBatch()}

	return d, nil
}

func (db *DB) Close() {
	db.db.Close()
}
