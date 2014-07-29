package ledis

import (
	"github.com/siddontang/ledisdb/store"
	"sync"
)

type tx struct {
	m sync.Mutex

	l  *Ledis
	wb *store.WriteBatch

	binlog *BinLog
	batch  [][]byte
}

func newTx(l *Ledis) *tx {
	t := new(tx)

	t.l = l
	t.wb = l.ldb.NewWriteBatch()

	t.batch = make([][]byte, 0, 4)
	t.binlog = l.binlog
	return t
}

func (t *tx) Close() {
	t.wb = nil
}

func (t *tx) Put(key []byte, value []byte) {
	t.wb.Put(key, value)

	if t.binlog != nil {
		buf := encodeBinLogPut(key, value)
		t.batch = append(t.batch, buf)
	}
}

func (t *tx) Delete(key []byte) {
	t.wb.Delete(key)

	if t.binlog != nil {
		buf := encodeBinLogDelete(key)
		t.batch = append(t.batch, buf)
	}
}

func (t *tx) Lock() {
	t.m.Lock()
}

func (t *tx) Unlock() {
	t.batch = t.batch[0:0]
	t.wb.Rollback()
	t.m.Unlock()
}

func (t *tx) Commit() error {
	var err error
	if t.binlog != nil {
		t.l.Lock()
		err = t.wb.Commit()
		if err != nil {
			t.l.Unlock()
			return err
		}

		err = t.binlog.Log(t.batch...)

		t.l.Unlock()
	} else {
		t.l.Lock()
		err = t.wb.Commit()
		t.l.Unlock()
	}
	return err
}

func (t *tx) Rollback() {
	t.wb.Rollback()
}
