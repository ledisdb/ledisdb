package ledis

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-leveldb/leveldb"
)

type Config struct {
	DataDB leveldb.Config `json:"data_db"`

	BinLog BinLogConfig `json:"binlog"`
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

	binlog *BinLog
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

	if len(cfg.BinLog.Path) > 0 {
		l.binlog, err = NewBinLogWithConfig(&cfg.BinLog)
		if err != nil {
			return nil, err
		}
	} else {
		l.binlog = nil
	}

	for i := uint8(0); i < MaxDBNumber; i++ {
		l.dbs[i] = newDB(l, i)
	}

	return l, nil
}

func newDB(l *Ledis, index uint8) *DB {
	d := new(DB)

	d.db = l.ldb

	d.index = index

	d.kvTx = newTx(l)
	d.listTx = newTx(l)
	d.hashTx = newTx(l)
	d.zsetTx = newTx(l)

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

func (l *Ledis) Snapshot() (*leveldb.Snapshot, string, int64) {
	if l.binlog == nil {
		return l.ldb.NewSnapshot(), "", 0
	} else {
		l.binlog.Lock()
		s := l.ldb.NewSnapshot()
		fileName, offset := l.binlog.SavePoint()
		l.binlog.Unlock()

		return s, fileName, offset
	}
}
