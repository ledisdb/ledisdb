package mdb

type Write struct {
	Key   []byte
	Value []byte
}

type WriteBatch struct {
	db *MDB
	wb []Write
}

func (w *WriteBatch) Close() error {
	return nil
}

func (w *WriteBatch) Put(key, value []byte) {
	w.wb = append(w.wb, Write{key, value})
}

func (w *WriteBatch) Delete(key []byte) {
	w.wb = append(w.wb, Write{key, nil})
}

func (w *WriteBatch) Commit() error {
	return w.db.BatchPut(w.wb)
}

func (w *WriteBatch) Rollback() error {
	w.wb = w.wb[0:0]
	return nil
}
