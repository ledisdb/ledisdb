package ssdb

import (
	"github.com/siddontang/golib/leveldb"
	"sync"
)

type tx struct {
	m sync.Mutex

	app *App

	wb *leveldb.WriteBatch
}

func (app *App) newTx() *tx {
	t := new(tx)

	t.app = app

	t.wb = app.db.NewWriteBatch()

	return t
}

func (t *tx) Close() {
	t.wb.Close()
}

func (t *tx) Put(key []byte, value []byte) {
	t.wb.Put(key, value)
}

func (t *tx) Delete(key []byte) {
	t.wb.Delete(key)
}

func (t *tx) Lock() {
	t.m.Lock()
}

func (t *tx) Unlock() {
	t.wb.Rollback()
	t.m.Unlock()
}

func (t *tx) Commit() error {
	err := t.wb.Commit()
	return err
}

func (t *tx) Rollback() {
	t.wb.Rollback()
}
