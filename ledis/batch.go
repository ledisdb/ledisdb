package ledis

import (
	"github.com/siddontang/ledisdb/store"
	"sync"
)

type batch struct {
	l *Ledis

	store.WriteBatch

	sync.Locker

	eb *eventBatch

	tx *Tx

	noLogging bool
}

func (b *batch) Commit() error {
	if b.l.replMode {
		return ErrWriteInReplMode
	}

	b.l.commitLock.Lock()
	defer b.l.commitLock.Unlock()

	if b.LogEanbled() {

	}

	err := b.WriteBatch.Commit()

	return err
}

// only use in expire cycle
func (b *batch) expireCommit() error {
	b.l.commitLock.Lock()
	defer b.l.commitLock.Unlock()

	return b.WriteBatch.Commit()
}

func (b *batch) Lock() {
	b.Locker.Lock()
}

func (b *batch) Unlock() {
	b.noLogging = false
	b.eb.Reset()
	b.WriteBatch.Rollback()
	b.Locker.Unlock()
}

func (b *batch) Put(key []byte, value []byte) {
	if b.LogEanbled() {
		b.eb.Put(key, value)
	}
	b.WriteBatch.Put(key, value)
}

func (b *batch) Delete(key []byte) {
	if b.LogEanbled() {
		b.eb.Delete(key)
	}

	b.WriteBatch.Delete(key)
}

func (b *batch) LogEanbled() bool {
	return !b.noLogging && b.l.log != nil
}

// only for expire cycle
func (b *batch) disableLog() {
	b.noLogging = true
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

type txBatchLocker struct {
}

func (l *txBatchLocker) Lock()   {}
func (l *txBatchLocker) Unlock() {}

type multiBatchLocker struct {
}

func (l *multiBatchLocker) Lock()   {}
func (l *multiBatchLocker) Unlock() {}

func (l *Ledis) newBatch(wb store.WriteBatch, locker sync.Locker, tx *Tx) *batch {
	b := new(batch)
	b.l = l
	b.WriteBatch = wb

	b.tx = tx
	b.Locker = locker

	if tx != nil {
		b.eb = tx.eb
	} else {
		b.eb = new(eventBatch)
	}
	b.noLogging = false

	return b
}
