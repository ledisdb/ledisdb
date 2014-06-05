package ledis

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-leveldb/leveldb"
	"github.com/siddontang/ledisdb/replication"
	"sync"
)

type Config struct {
	DataDB leveldb.Config `json:"data_db"`

	BinLog   replication.BinLogConfig   `json:"binlog"`
	RelayLog replication.RelayLogConfig `json:"relaylog"`
}

type DB struct {
	l *Ledis

	db *leveldb.DB

	index uint8

	kvTx   *tx
	listTx *tx
	hashTx *tx
	zsetTx *tx
}

type Ledis struct {
	sync.Mutex

	cfg *Config

	ldb *leveldb.DB
	dbs [MaxDBNumber]*DB

	binlog   *replication.Log
	relaylog *replication.Log

	quit chan struct{}
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

	l.quit = make(chan struct{})

	l.ldb = ldb

	if len(cfg.BinLog.Path) > 0 {
		l.binlog, err = replication.NewBinLogWithConfig(&cfg.BinLog)
		if err != nil {
			return nil, err
		}
	} else {
		l.binlog = nil
	}

	if len(cfg.RelayLog.Path) > 0 {
		l.relaylog, err = replication.NewRelayLogWithConfig(&cfg.RelayLog)
		if err != nil {
			return nil, err
		}
	} else {
		l.relaylog = nil
	}

	for i := uint8(0); i < MaxDBNumber; i++ {
		l.dbs[i] = newDB(l, i)
	}

	return l, nil
}

func newDB(l *Ledis, index uint8) *DB {
	d := new(DB)

	d.l = l

	d.db = l.ldb

	d.index = index

	d.kvTx = newTx(l)
	d.listTx = newTx(l)
	d.hashTx = newTx(l)
	d.zsetTx = newTx(l)

	d.activeExpireCycle()

	return d
}

func (l *Ledis) Close() {
	close(l.quit)

	l.ldb.Close()

	if l.binlog != nil {
		l.binlog.Close()
		l.binlog = nil
	}

	if l.relaylog != nil {
		l.relaylog.Close()
		l.relaylog = nil
	}
}

func (l *Ledis) Select(index int) (*DB, error) {
	if index < 0 || index >= int(MaxDBNumber) {
		return nil, fmt.Errorf("invalid db index %d", index)
	}

	return l.dbs[index], nil
}
