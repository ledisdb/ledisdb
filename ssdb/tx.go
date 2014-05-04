package ssdb

import (
	"github.com/siddontang/golib/leveldb"
)

type tx struct {
	app *App

	wb *leveldb.WriteBatch
}

func (app *App) newTx() *tx {
	t := new(tx)

	t.app = app

	t.wb = app.db.NewWriteBatch()

	return t
}

func (t *tx) Put(key []byte, value []byte) {
	t.wb.Put(key, value)
}

func (t *tx) Delete(key []byte) {
	t.wb.Delete(key)
}

func (t *tx) Commit() error {
	err := t.wb.Commit()
	return err
}

func (t *tx) Rollback() error {
	err := t.wb.Rollback()
	return err
}
