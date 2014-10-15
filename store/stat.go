package store

import (
	"github.com/siddontang/go/sync2"
)

type Stat struct {
	GetNum             sync2.AtomicInt64
	GetMissingNum      sync2.AtomicInt64
	PutNum             sync2.AtomicInt64
	DeleteNum          sync2.AtomicInt64
	SyncPutNum         sync2.AtomicInt64
	SyncDeleteNum      sync2.AtomicInt64
	IterNum            sync2.AtomicInt64
	IterSeekNum        sync2.AtomicInt64
	IterCloseNum       sync2.AtomicInt64
	SnapshotNum        sync2.AtomicInt64
	SnapshotCloseNum   sync2.AtomicInt64
	BatchNum           sync2.AtomicInt64
	BatchCommitNum     sync2.AtomicInt64
	BatchSyncCommitNum sync2.AtomicInt64
	TxNum              sync2.AtomicInt64
	TxCommitNum        sync2.AtomicInt64
	TxCloseNum         sync2.AtomicInt64
	CompactNum         sync2.AtomicInt64
	CompactTotalTime   sync2.AtomicDuration
}

func (st *Stat) statGet(v []byte, err error) {
	st.GetNum.Add(1)
	if v == nil && err == nil {
		st.GetMissingNum.Add(1)
	}
}

func (st *Stat) Reset() {
	*st = Stat{}
	// st.GetNum.Set(0)
	// st.GetMissingNum.Set(0)
	// st.PutNum.Set(0)
	// st.DeleteNum.Set(0)
	// st.SyncPutNum.Set(0)
	// st.SyncDeleteNum.Set(0)
	// st.IterNum.Set(0)
	// st.IterSeekNum.Set(0)
	// st.IterCloseNum.Set(0)
	// st.SnapshotNum.Set(0)
	// st.SnapshotCloseNum.Set(0)
	// st.BatchNum.Set(0)
	// st.BatchCommitNum.Set(0)
	// st.BatchSyncCommitNum.Set(0)
	// st.TxNum.Set(0)
	// st.TxCommitNum.Set(0)
	// st.TxCloseNum.Set(0)
	// st.CompactNum.Set(0)
	// st.CompactTotalTime.Set(0)
}
