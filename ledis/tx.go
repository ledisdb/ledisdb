package ledis

import (
	"errors"
	"github.com/siddontang/ledisdb/store"
	"sync"
)

var (
	ErrNestTx = errors.New("nest transaction not supported")
	ErrTxDone = errors.New("Transaction has already been committed or rolled back")
)

type batch struct {
	l *Ledis

	store.WriteBatch

	sync.Locker

	logs [][]byte

	tx *Tx
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

func (l *Ledis) newBatch(wb store.WriteBatch, tx *Tx) *batch {
	b := new(batch)
	b.l = l
	b.WriteBatch = wb

	b.tx = tx
	if tx == nil {
		b.Locker = &dbBatchLocker{l: &sync.Mutex{}, wrLock: &l.wLock}
	} else {
		b.Locker = &txBatchLocker{}
	}

	b.logs = [][]byte{}
	return b
}

func (db *DB) newBatch() *batch {
	return db.l.newBatch(db.bucket.NewWriteBatch(), nil)
}

func (b *batch) Commit() error {
	b.l.commitLock.Lock()
	defer b.l.commitLock.Unlock()

	err := b.WriteBatch.Commit()

	if b.l.binlog != nil {
		if err == nil {
			if b.tx == nil {
				b.l.binlog.Log(b.logs...)
			} else {
				b.tx.logs = append(b.tx.logs, b.logs...)
			}
		}
		b.logs = [][]byte{}
	}

	return err
}

func (b *batch) Lock() {
	b.Locker.Lock()
}

func (b *batch) Unlock() {
	if b.l.binlog != nil {
		b.logs = [][]byte{}
	}
	b.WriteBatch.Rollback()
	b.Locker.Unlock()
}

func (b *batch) Put(key []byte, value []byte) {
	if b.l.binlog != nil {
		buf := encodeBinLogPut(key, value)
		b.logs = append(b.logs, buf)
	}
	b.WriteBatch.Put(key, value)
}

func (b *batch) Delete(key []byte) {
	if b.l.binlog != nil {
		buf := encodeBinLogDelete(key)
		b.logs = append(b.logs, buf)
	}
	b.WriteBatch.Delete(key)
}

type Tx struct {
	*DB

	tx *store.Tx

	logs [][]byte

	index uint8
}

func (db *DB) IsTransaction() bool {
	return db.isTx
}

// Begin a transaction, it will block all other write operations before calling Commit or Rollback.
// You must be very careful to prevent long-time transaction.
func (db *DB) Begin() (*Tx, error) {
	if db.isTx {
		return nil, ErrNestTx
	}

	tx := new(Tx)

	tx.DB = new(DB)
	tx.DB.l = db.l

	tx.l.wLock.Lock()

	tx.index = db.index

	tx.DB.sdb = db.sdb

	var err error
	tx.tx, err = db.sdb.Begin()
	if err != nil {
		tx.l.wLock.Unlock()
		return nil, err
	}

	tx.DB.bucket = tx.tx

	tx.DB.isTx = true

	tx.DB.index = db.index

	tx.DB.kvBatch = tx.newTxBatch()
	tx.DB.listBatch = tx.newTxBatch()
	tx.DB.hashBatch = tx.newTxBatch()
	tx.DB.zsetBatch = tx.newTxBatch()
	tx.DB.binBatch = tx.newTxBatch()
	tx.DB.setBatch = tx.newTxBatch()

	return tx, nil
}

func (tx *Tx) Commit() error {
	if tx.tx == nil {
		return ErrTxDone
	}

	tx.l.commitLock.Lock()
	err := tx.tx.Commit()
	tx.tx = nil

	if len(tx.logs) > 0 {
		tx.l.binlog.Log(tx.logs...)
	}

	tx.l.commitLock.Unlock()

	tx.l.wLock.Unlock()
	tx.DB = nil
	return err
}

func (tx *Tx) Rollback() error {
	if tx.tx == nil {
		return ErrTxDone
	}

	err := tx.tx.Rollback()
	tx.tx = nil

	tx.l.wLock.Unlock()
	tx.DB = nil
	return err
}

func (tx *Tx) newTxBatch() *batch {
	return tx.l.newBatch(tx.tx.NewWriteBatch(), tx)
}

func (tx *Tx) Index() int {
	return int(tx.index)
}
