package rpl

import (
	"encoding/binary"
	"github.com/siddontang/go/log"
	"github.com/siddontang/go/snappy"
	"github.com/siddontang/ledisdb/config"
	"os"
	"path"
	"sync"
	"time"
)

type Stat struct {
	FirstID  uint64
	LastID   uint64
	CommitID uint64
}

type Replication struct {
	m sync.Mutex

	cfg *config.Config

	s LogStore

	commitID  uint64
	commitLog *os.File

	quit chan struct{}

	wg sync.WaitGroup

	nc chan struct{}

	ncm sync.Mutex
}

func NewReplication(cfg *config.Config) (*Replication, error) {
	if len(cfg.Replication.Path) == 0 {
		cfg.Replication.Path = path.Join(cfg.DataDir, "rpl")
	}

	base := cfg.Replication.Path

	r := new(Replication)

	r.quit = make(chan struct{})
	r.nc = make(chan struct{})

	r.cfg = cfg

	var err error
	if r.s, err = NewGoLevelDBStore(path.Join(base, "wal"), cfg.Replication.SyncLog); err != nil {
		return nil, err
	}

	if r.commitLog, err = os.OpenFile(path.Join(base, "commit.log"), os.O_RDWR|os.O_CREATE, 0644); err != nil {
		return nil, err
	}

	if s, _ := r.commitLog.Stat(); s.Size() == 0 {
		r.commitID = 0
	} else if err = binary.Read(r.commitLog, binary.BigEndian, &r.commitID); err != nil {
		return nil, err
	}

	r.wg.Add(1)
	go r.onPurgeExpired()

	return r, nil
}

func (r *Replication) Close() error {
	close(r.quit)

	r.wg.Wait()

	r.m.Lock()
	defer r.m.Unlock()

	if r.s != nil {
		r.s.Close()
		r.s = nil
	}

	if r.commitLog != nil {
		r.commitLog.Close()
		r.commitLog = nil
	}

	return nil
}

func (r *Replication) Log(data []byte) (*Log, error) {
	if r.cfg.Replication.Compression {
		//todo optimize
		var err error
		if data, err = snappy.Encode(nil, data); err != nil {
			return nil, err
		}
	}

	r.m.Lock()
	defer r.m.Unlock()

	lastID, err := r.s.LastID()
	if err != nil {
		return nil, err
	}

	commitId := r.commitID
	if lastID < commitId {
		lastID = commitId
	}

	l := new(Log)
	l.ID = lastID + 1
	l.CreateTime = uint32(time.Now().Unix())

	if r.cfg.Replication.Compression {
		l.Compression = 1
	} else {
		l.Compression = 0
	}

	l.Data = data

	if err = r.s.StoreLog(l); err != nil {
		return nil, err
	}

	r.ncm.Lock()
	close(r.nc)
	r.nc = make(chan struct{})
	r.ncm.Unlock()

	return l, nil
}

func (r *Replication) WaitLog() <-chan struct{} {
	r.ncm.Lock()
	ch := r.nc
	r.ncm.Unlock()
	return ch
}

func (r *Replication) StoreLog(log *Log) error {
	return r.StoreLogs([]*Log{log})
}

func (r *Replication) StoreLogs(logs []*Log) error {
	r.m.Lock()
	defer r.m.Unlock()

	return r.s.StoreLogs(logs)
}

func (r *Replication) FirstLogID() (uint64, error) {
	r.m.Lock()
	defer r.m.Unlock()
	id, err := r.s.FirstID()
	return id, err
}

func (r *Replication) LastLogID() (uint64, error) {
	r.m.Lock()
	defer r.m.Unlock()
	id, err := r.s.LastID()
	return id, err
}

func (r *Replication) LastCommitID() (uint64, error) {
	r.m.Lock()
	id := r.commitID
	r.m.Unlock()
	return id, nil
}

func (r *Replication) UpdateCommitID(id uint64) error {
	r.m.Lock()
	defer r.m.Unlock()

	return r.updateCommitID(id)
}

func (r *Replication) Stat() (*Stat, error) {
	r.m.Lock()
	defer r.m.Unlock()

	s := &Stat{}
	var err error

	if s.FirstID, err = r.s.FirstID(); err != nil {
		return nil, err
	}

	if s.LastID, err = r.s.LastID(); err != nil {
		return nil, err
	}

	s.CommitID = r.commitID
	return s, nil
}

func (r *Replication) updateCommitID(id uint64) error {
	if _, err := r.commitLog.Seek(0, os.SEEK_SET); err != nil {
		return err
	}

	if err := binary.Write(r.commitLog, binary.BigEndian, id); err != nil {
		return err
	}

	r.commitID = id

	return nil
}

func (r *Replication) CommitIDBehind() (bool, error) {
	r.m.Lock()
	defer r.m.Unlock()

	id, err := r.s.LastID()
	if err != nil {
		return false, err
	}

	return id > r.commitID, nil
}

func (r *Replication) GetLog(id uint64, log *Log) error {
	return r.s.GetLog(id, log)
}

func (r *Replication) NextNeedCommitLog(log *Log) error {
	r.m.Lock()
	defer r.m.Unlock()

	id, err := r.s.LastID()
	if err != nil {
		return err
	}

	if id <= r.commitID {
		return ErrNoBehindLog
	}

	return r.s.GetLog(r.commitID+1, log)

}

func (r *Replication) Clear() error {
	return r.ClearWithCommitID(0)
}

func (r *Replication) ClearWithCommitID(id uint64) error {
	r.m.Lock()
	defer r.m.Unlock()

	if err := r.s.Clear(); err != nil {
		return err
	}

	return r.updateCommitID(id)
}

func (r *Replication) onPurgeExpired() {
	defer r.wg.Done()

	for {
		select {
		case <-time.After(1 * time.Hour):
			n := (r.cfg.Replication.ExpiredLogDays * 24 * 3600)
			r.m.Lock()
			if err := r.s.PurgeExpired(int64(n)); err != nil {
				log.Error("purge expired log error %s", err.Error())
			}
			r.m.Unlock()
		case <-r.quit:
			return
		}
	}
}
