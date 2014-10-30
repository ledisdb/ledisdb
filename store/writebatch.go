package store

import (
	"github.com/siddontang/ledisdb/store/driver"
	"time"
)

type WriteBatch struct {
	wb driver.IWriteBatch
	st *Stat

	db *DB
}

func (wb *WriteBatch) Put(key []byte, value []byte) {
	wb.wb.Put(key, value)
}

func (wb *WriteBatch) Delete(key []byte) {
	wb.wb.Delete(key)
}

func (wb *WriteBatch) Commit() error {
	wb.st.BatchCommitNum.Add(1)
	var err error
	t := time.Now()
	if wb.db == nil || !wb.db.needSyncCommit() {
		err = wb.wb.Commit()
	} else {
		err = wb.wb.SyncCommit()
	}

	wb.st.BatchCommitTotalTime.Add(time.Now().Sub(t))

	return err
}

func (wb *WriteBatch) Rollback() error {
	return wb.wb.Rollback()
}
