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
	sync.Mutex
	dbLock *sync.RWMutex
}

type txBatchLocker struct {
}

func (l *txBatchLocker) Lock() {
}

func (l *txBatchLocker) Unlock() {
}

func (l *dbBatchLocker) Lock() {
	l.dbLock.RLock()
	l.Mutex.Lock()
}

func (l *dbBatchLocker) Unlock() {
	l.Mutex.Unlock()
	l.dbLock.RUnlock()
}

func (db *DB) newBatch() *batch {
	b := new(batch)

	b.WriteBatch = db.bucket.NewWriteBatch()
	b.Locker = &dbBatchLocker{dbLock: db.dbLock}
	b.l = db.l

	return b
}

func (b *batch) Commit() error {
	b.l.Lock()
	defer b.l.Unlock()

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
	b.Rollback()
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
	tx.DB.dbLock = db.dbLock

	tx.DB.dbLock.Lock()

	tx.DB.l = db.l

	tx.DB.sdb = db.sdb

	var err error
	tx.tx, err = db.sdb.Begin()
	if err != nil {
		tx.DB.dbLock.Unlock()
		return nil, err
	}

	tx.DB.bucket = tx.tx

	tx.DB.isTx = true

	tx.DB.index = db.index

	tx.DB.kvBatch = tx.newBatch()
	tx.DB.listBatch = tx.newBatch()
	tx.DB.hashBatch = tx.newBatch()
	tx.DB.zsetBatch = tx.newBatch()
	tx.DB.binBatch = tx.newBatch()
	tx.DB.setBatch = tx.newBatch()

	return tx, nil
}

func (tx *Tx) Commit() error {
	if tx.tx == nil {
		return ErrTxDone
	}

	tx.l.Lock()
	err := tx.tx.Commit()
	tx.tx = nil

	if len(tx.logs) > 0 {
		tx.l.binlog.Log(tx.logs...)
	}

	tx.l.Unlock()

	tx.DB.dbLock.Unlock()
	tx.DB = nil
	return err
}

func (tx *Tx) Rollback() error {
	if tx.tx == nil {
		return ErrTxDone
	}

	err := tx.tx.Rollback()
	tx.tx = nil

	tx.DB.dbLock.Unlock()
	tx.DB = nil
	return err
}

func (tx *Tx) newBatch() *batch {
	b := new(batch)

	b.l = tx.l
	b.WriteBatch = tx.tx.NewWriteBatch()
	b.Locker = &txBatchLocker{}
	b.tx = tx

	return b
}
