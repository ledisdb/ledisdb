package ledis

import (
	"encoding/binary"
	"github.com/siddontang/go-leveldb/leveldb"
	"sync"
)

type tx struct {
	m sync.Mutex

	wb *leveldb.WriteBatch

	binlog *BinLog
	batch  [][]byte
}

func newTx(l *Ledis) *tx {
	t := new(tx)

	t.wb = l.ldb.NewWriteBatch()

	t.batch = make([][]byte, 0, 4)
	t.binlog = l.binlog
	return t
}

func (t *tx) Close() {
	t.wb.Close()
}

func (t *tx) Put(key []byte, value []byte) {
	t.wb.Put(key, value)

	if t.binlog != nil {
		buf := make([]byte, 7+len(key)+len(value))
		buf[0] = BinLogTypeValue
		pos := 1
		binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
		pos += 2
		copy(buf[pos:], key)
		pos += len(key)
		binary.BigEndian.PutUint32(buf[pos:], uint32(len(value)))
		pos += 4
		copy(buf[pos:], value)

		t.batch = append(t.batch, buf)
	}
}

func (t *tx) Delete(key []byte) {
	t.wb.Delete(key)

	if t.binlog != nil {
		buf := make([]byte, 3+len(key))
		buf[0] = BinLogTypeDeletion
		pos := 1
		binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
		pos += 2
		copy(buf[pos:], key)

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
		t.binlog.Lock()
		err = t.wb.Commit()
		if err != nil {
			t.binlog.Unlock()
			return err
		}

		err = t.binlog.Log(t.batch...)

		t.binlog.Unlock()
	} else {
		err = t.wb.Commit()
	}
	return err
}

func (t *tx) Rollback() {
	t.wb.Rollback()
}
