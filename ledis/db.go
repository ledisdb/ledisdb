package ledis

import (
	"encoding/json"
	"github.com/siddontang/go-leveldb/leveldb"
)

type DBConfig struct {
	DataDB leveldb.Config `json:"data_db"`
}

type DB struct {
	cfg *DBConfig

	db *leveldb.DB

	kvTx   *tx
	listTx *tx
	hashTx *tx
	zsetTx *tx
}

func OpenDB(configJson json.RawMessage) (*DB, error) {
	var cfg DBConfig

	if err := json.Unmarshal(configJson, &cfg); err != nil {
		return nil, err
	}

	return OpenDBWithConfig(&cfg)
}

func OpenDBWithConfig(cfg *DBConfig) (*DB, error) {
	db, err := leveldb.OpenWithConfig(&cfg.DataDB)
	if err != nil {
		return nil, err
	}

	d := new(DB)

	d.cfg = cfg

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
