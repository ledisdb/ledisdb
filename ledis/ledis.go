package ledis

import (
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/rpl"
	"github.com/siddontang/ledisdb/store"
	"sync"
	"time"
)

type Ledis struct {
	cfg *config.Config

	ldb *store.DB
	dbs [MaxDBNumber]*DB

	quit chan struct{}
	wg   sync.WaitGroup

	r      *rpl.Replication
	rc     chan struct{}
	rbatch store.WriteBatch
	rwg    sync.WaitGroup

	wLock      sync.RWMutex //allow one write at same time
	commitLock sync.Mutex   //allow one write commit at same time

	// for readonly mode, only replication can write
	readOnly bool
}

func Open(cfg *config.Config) (*Ledis, error) {
	return Open2(cfg, RDWRMode)
}

func Open2(cfg *config.Config, flags int) (*Ledis, error) {
	if len(cfg.DataDir) == 0 {
		cfg.DataDir = config.DefaultDataDir
	}

	ldb, err := store.Open(cfg)
	if err != nil {
		return nil, err
	}

	l := new(Ledis)

	l.readOnly = (flags&ROnlyMode > 0)

	l.quit = make(chan struct{})

	l.ldb = ldb

	if cfg.Replication.Use {
		if l.r, err = rpl.NewReplication(cfg); err != nil {
			return nil, err
		}

		l.rc = make(chan struct{})
		l.rbatch = l.ldb.NewWriteBatch()

		go l.onReplication()
	} else {
		l.r = nil
	}

	for i := uint8(0); i < MaxDBNumber; i++ {
		l.dbs[i] = l.newDB(i)
	}

	go l.onDataExpired()

	return l, nil
}

func (l *Ledis) Close() {
	close(l.quit)
	l.wg.Wait()

	l.ldb.Close()

	if l.r != nil {
		l.r.Close()
		l.r = nil
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

func (l *Ledis) IsReadOnly() bool {
	if l.readOnly {
		return true
	} else if l.r != nil {
		if b, _ := l.r.CommitIDBehind(); b {
			return true
		}
	}
	return false
}

func (l *Ledis) SetReadOnly(b bool) {
	l.readOnly = b
}

func (l *Ledis) onDataExpired() {
	l.wg.Add(1)
	defer l.wg.Done()

	var executors []*elimination = make([]*elimination, len(l.dbs))
	for i, db := range l.dbs {
		executors[i] = db.newEliminator()
	}

	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	done := make(chan struct{})

	for {
		select {
		case <-tick.C:
			if l.IsReadOnly() {
				break
			}

			go func() {
				for _, eli := range executors {
					eli.active()
				}
				done <- struct{}{}
			}()
			<-done
		case <-l.quit:
			return
		}
	}

}
