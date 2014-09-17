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
	b.l.commitLock.Lock()
	defer b.l.commitLock.Unlock()

	err := b.WriteBatch.Commit()

	return err
}

func (b *batch) Lock() {
	b.Locker.Lock()
}

func (b *batch) Unlock() {
	b.noLogging = false
	b.WriteBatch.Rollback()
	b.Locker.Unlock()
}

func (b *batch) Put(key []byte, value []byte) {
	b.WriteBatch.Put(key, value)
}

func (b *batch) Delete(key []byte) {

	b.WriteBatch.Delete(key)
}

func (b *batch) LogEanbled() bool {
	return !b.noLogging && b.l.log != nil
}

func (b *batch) DisableLog(d bool) {
	b.noLogging = d
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

	b.eb = new(eventBatch)
	b.noLogging = false

	return b
}
