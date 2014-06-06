package ledis

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-leveldb/leveldb"
	"github.com/siddontang/ledisdb/replication"
	"path"
	"sync"
)

type Config struct {
	DataDir string `json:"data_dir"`

	//data_db path is data_dir/data
	//if you not set leveldb path, use data_dir/data
	DataDB leveldb.Config `json:"data_db"`

	//binlog path is data_dir/binlog
	//you muse set binlog name to enable binlog
	//if you not set bin log path, use data_dir/bin_log
	BinLog replication.BinLogConfig `json:"bin_log"`
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

	binlog *replication.Log

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
	if len(cfg.DataDir) == 0 {
		return nil, fmt.Errorf("must set correct data_dir")
	}

	if len(cfg.DataDB.Path) == 0 {
		cfg.DataDB.Path = path.Join(cfg.DataDir, "data")
	}

	ldb, err := leveldb.OpenWithConfig(&cfg.DataDB)
	if err != nil {
		return nil, err
	}

	l := new(Ledis)

	l.quit = make(chan struct{})

	l.ldb = ldb

	if len(cfg.BinLog.Name) > 0 {
		if len(cfg.BinLog.Path) == 0 {
			cfg.BinLog.Path = path.Join(cfg.DataDir, "bin_log")
		}
		l.binlog, err = replication.NewBinLogWithConfig(&cfg.BinLog)
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
}

func (l *Ledis) Select(index int) (*DB, error) {
	if index < 0 || index >= int(MaxDBNumber) {
		return nil, fmt.Errorf("invalid db index %d", index)
	}

	return l.dbs[index], nil
}
