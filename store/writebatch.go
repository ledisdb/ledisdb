package store

import (
	"github.com/siddontang/ledisdb/store/driver"
)

type WriteBatch struct {
	wb        driver.IWriteBatch
	st        *Stat
	putNum    int64
	deleteNum int64

	db *DB
}

func (wb *WriteBatch) Put(key []byte, value []byte) {
	wb.putNum++
	wb.wb.Put(key, value)
}

func (wb *WriteBatch) Delete(key []byte) {
	wb.deleteNum++
	wb.wb.Delete(key)
}

func (wb *WriteBatch) Commit() error {
	wb.st.BatchCommitNum.Add(1)
	wb.st.PutNum.Add(wb.putNum)
	wb.st.DeleteNum.Add(wb.deleteNum)
	wb.putNum = 0
	wb.deleteNum = 0
	if wb.db == nil || !wb.db.needSyncCommit() {
		return wb.wb.Commit()
	} else {
		return wb.wb.SyncCommit()
	}
}

func (wb *WriteBatch) Rollback() error {
	return wb.wb.Rollback()
}
