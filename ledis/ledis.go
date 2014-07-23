package ledis

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/leveldb"
	"sync"
	"time"
)

type DB struct {
	l *Ledis

	db *leveldb.DB

	index uint8

	kvTx   *tx
	listTx *tx
	hashTx *tx
	zsetTx *tx
	binTx  *tx
}

type Ledis struct {
	sync.Mutex

	cfg *Config

	ldb *leveldb.DB
	dbs [MaxDBNumber]*DB

	binlog *BinLog

	quit chan struct{}
	jobs *sync.WaitGroup
}

func OpenWithJsonConfig(configJson json.RawMessage) (*Ledis, error) {
	var cfg Config

	if err := json.Unmarshal(configJson, &cfg); err != nil {
		return nil, err
	}

	return Open(&cfg)
}

func Open(cfg *Config) (*Ledis, error) {
	if len(cfg.DataDir) == 0 {
		return nil, fmt.Errorf("must set correct data_dir")
	}

	ldb, err := leveldb.Open(cfg.NewDBConfig())
	if err != nil {
		return nil, err
	}

	l := new(Ledis)

	l.quit = make(chan struct{})
	l.jobs = new(sync.WaitGroup)

	l.ldb = ldb

	if cfg.BinLog.Use {
		l.binlog, err = NewBinLog(cfg.NewBinLogConfig())
		if err != nil {
			return nil, err
		}
	} else {
		l.binlog = nil
	}

	for i := uint8(0); i < MaxDBNumber; i++ {
		l.dbs[i] = newDB(l, i)
	}

	l.activeExpireCycle()

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
	d.binTx = newTx(l)

	return d
}

func (l *Ledis) Close() {
	close(l.quit)
	l.jobs.Wait()

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

func (l *Ledis) FlushAll() error {
	for index, db := range l.dbs {
		if _, err := db.FlushAll(); err != nil {
			log.Error("flush db %d error %s", index, err.Error())
		}
	}

	return nil
}

// very dangerous to use
func (l *Ledis) DataDB() *leveldb.DB {
	return l.ldb
}

func (l *Ledis) activeExpireCycle() {
	var executors []*elimination = make([]*elimination, len(l.dbs))
	for i, db := range l.dbs {
		executors[i] = db.newEliminator()
	}

	l.jobs.Add(1)
	go func() {
		tick := time.NewTicker(1 * time.Second)
		end := false
		done := make(chan struct{})
		for !end {
			select {
			case <-tick.C:
				go func() {
					for _, eli := range executors {
						eli.active()
					}
					done <- struct{}{}
				}()
				<-done
			case <-l.quit:
				end = true
				break
			}
		}

		tick.Stop()
		l.jobs.Done()
	}()
}
