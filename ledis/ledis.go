package ledis

import (
	"fmt"
	"github.com/siddontang/go/filelock"
	"github.com/siddontang/go/log"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/rpl"
	"github.com/siddontang/ledisdb/store"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

type Ledis struct {
	cfg *config.Config

	ldb *store.DB
	dbs [MaxDBNumber]*DB

	quit chan struct{}
	wg   sync.WaitGroup

	//for replication
	r      *rpl.Replication
	rc     chan struct{}
	rbatch store.WriteBatch
	rwg    sync.WaitGroup
	rhs    []NewLogEventHandler

	wLock      sync.RWMutex //allow one write at same time
	commitLock sync.Mutex   //allow one write commit at same time

	// for readonly mode, only replication can write
	readOnly bool

	lock io.Closer
}

func Open(cfg *config.Config) (*Ledis, error) {
	return Open2(cfg, RDWRMode)
}

func Open2(cfg *config.Config, flags int) (*Ledis, error) {
	if len(cfg.DataDir) == 0 {
		cfg.DataDir = config.DefaultDataDir
	}

	os.MkdirAll(cfg.DataDir, 0755)

	var err error

	l := new(Ledis)

	if l.lock, err = filelock.Lock(path.Join(cfg.DataDir, "LOCK")); err != nil {
		return nil, err
	}

	l.readOnly = (flags&ROnlyMode > 0)

	l.quit = make(chan struct{})

	if l.ldb, err = store.Open(cfg); err != nil {
		return nil, err
	}

	if cfg.UseReplication {
		if l.r, err = rpl.NewReplication(cfg); err != nil {
			return nil, err
		}

		l.rc = make(chan struct{}, 8)
		l.rbatch = l.ldb.NewWriteBatch()

		go l.onReplication()

		//first we must try wait all replication ok
		//maybe some logs are not committed
		l.WaitReplication()
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

	if l.lock != nil {
		l.lock.Close()
		l.lock = nil
	}
}

func (l *Ledis) Select(index int) (*DB, error) {
	if index < 0 || index >= int(MaxDBNumber) {
		return nil, fmt.Errorf("invalid db index %d", index)
	}

	return l.dbs[index], nil
}

// Flush All will clear all data and replication logs
func (l *Ledis) FlushAll() error {
	l.wLock.Lock()
	defer l.wLock.Unlock()

	return l.flushAll()
}

func (l *Ledis) flushAll() error {
	it := l.ldb.NewIterator()
	defer it.Close()

	w := l.ldb.NewWriteBatch()
	defer w.Rollback()

	n := 0
	for ; it.Valid(); it.Next() {
		n++
		if n == 10000 {
			if err := w.Commit(); err != nil {
				log.Fatal("flush all commit error: %s", err.Error())
				return err
			}
			n = 0
		}
		w.Delete(it.RawKey())
	}

	if err := w.Commit(); err != nil {
		log.Fatal("flush all commit error: %s", err.Error())
		return err
	}

	if l.r != nil {
		if err := l.r.Clear(); err != nil {
			log.Fatal("flush all replication clear error: %s", err.Error())
			return err
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
