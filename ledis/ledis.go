package ledis

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-leveldb/leveldb"
)

type Config struct {
	DataDB leveldb.Config `json:"data_db"`
}

type DB struct {
	db *leveldb.DB

	index uint8

	kvTx   *tx
	listTx *tx
	hashTx *tx
	zsetTx *tx
}

type Ledis struct {
	cfg *Config

	ldb *leveldb.DB
	dbs [MaxDBNumber]*DB
}

func Open(configJson json.RawMessage) (*Ledis, error) {
	var cfg Config

	if err := json.Unmarshal(configJson, &cfg); err != nil {
		return nil, err
	}

	return OpenWithConfig(&cfg)
}

func OpenWithConfig(cfg *Config) (*Ledis, error) {
	ldb, err := leveldb.OpenWithConfig(&cfg.DataDB)
	if err != nil {
		return nil, err
	}

	l := new(Ledis)
	l.ldb = ldb

	for i := uint8(0); i < MaxDBNumber; i++ {
		l.dbs[i] = newDB(l, i)
	}

	return l, nil
}

func newDB(l *Ledis, index uint8) *DB {
	d := new(DB)

	d.db = l.ldb

	d.index = index

	d.kvTx = &tx{wb: d.db.NewWriteBatch()}
	d.listTx = &tx{wb: d.db.NewWriteBatch()}
	d.hashTx = &tx{wb: d.db.NewWriteBatch()}
	d.zsetTx = &tx{wb: d.db.NewWriteBatch()}

	return d
}

func (l *Ledis) Close() {
	l.ldb.Close()
}

func (l *Ledis) Select(index int) (*DB, error) {
	if index < 0 || index >= int(MaxDBNumber) {
		return nil, fmt.Errorf("invalid db index %d", index)
	}

	return l.dbs[index], nil
}
