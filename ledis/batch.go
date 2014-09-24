package ledis

import (
	"github.com/siddontang/go/log"
	"github.com/siddontang/ledisdb/rpl"
	"github.com/siddontang/ledisdb/store"
	"sync"
)

type batch struct {
	l *Ledis

	store.WriteBatch

	sync.Locker

	eb *eventBatch
}

func (b *batch) Commit() error {
	if b.l.IsReadOnly() {
		return ErrWriteInROnly
	}

	b.l.commitLock.Lock()
	defer b.l.commitLock.Unlock()

	var err error
	if b.l.r != nil {
		var l *rpl.Log
		if l, err = b.l.r.Log(b.eb.Bytes()); err != nil {
			log.Fatal("write wal error %s", err.Error())
			return err
		}

		b.l.propagate(l)

		if err = b.WriteBatch.Commit(); err != nil {
			log.Fatal("commit error %s", err.Error())
			return err
		}

		if err = b.l.r.UpdateCommitID(l.ID); err != nil {
			log.Fatal("update commit id error %s", err.Error())
			return err
		}

		return nil
	} else {
		return b.WriteBatch.Commit()
	}
}

func (b *batch) Lock() {
	b.Locker.Lock()
}

func (b *batch) Unlock() {
	b.eb.Reset()

	b.WriteBatch.Rollback()
	b.Locker.Unlock()
}

func (b *batch) Put(key []byte, value []byte) {
	if b.l.r != nil {
		b.eb.Put(key, value)
	}

	b.WriteBatch.Put(key, value)
}

func (b *batch) Delete(key []byte) {
	if b.l.r != nil {
		b.Delete(key)
	}

	b.WriteBatch.Delete(key)
}

type dbBatchLocker struct {
	l      *sync.Mutex
	wrLock *sync.RWMutex
}

func (l *dbBatchLocker) Lock() {
	l.wrLock.RLock()
	l.l.Lock()
}

func (l *dbBatchLocker) Unlock() {
	l.l.Unlock()
	l.wrLock.RUnlock()
}

type multiBatchLocker struct {
}

func (l *multiBatchLocker) Lock()   {}
func (l *multiBatchLocker) Unlock() {}

func (l *Ledis) newBatch(wb store.WriteBatch, locker sync.Locker) *batch {
	b := new(batch)
	b.l = l
	b.WriteBatch = wb

	b.Locker = locker

	b.eb = new(eventBatch)

	return b
}
