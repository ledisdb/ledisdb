package store

import (
	"github.com/siddontang/ledisdb/store/driver"
)

type WriteBatch struct {
	driver.IWriteBatch
	st        *Stat
	putNum    int64
	deleteNum int64
}

func (wb *WriteBatch) Put(key []byte, value []byte) {
	wb.putNum++
	wb.IWriteBatch.Put(key, value)
}

func (wb *WriteBatch) Delete(key []byte) {
	wb.deleteNum++
	wb.IWriteBatch.Delete(key)
}

func (wb *WriteBatch) Commit() error {
	wb.st.BatchCommitNum.Add(1)
	wb.st.PutNum.Add(wb.putNum)
	wb.st.DeleteNum.Add(wb.deleteNum)
	wb.putNum = 0
	wb.deleteNum = 0
	return wb.IWriteBatch.Commit()
}

func (wb *WriteBatch) SyncCommit() error {
	wb.st.BatchSyncCommitNum.Add(1)
	wb.st.SyncPutNum.Add(wb.putNum)
	wb.st.SyncDeleteNum.Add(wb.deleteNum)
	wb.putNum = 0
	wb.deleteNum = 0
	return wb.IWriteBatch.SyncCommit()
}

func (wb *WriteBatch) Rollback() error {
	return wb.IWriteBatch.Rollback()
}
