package ledis

import (
	"github.com/siddontang/go/num"
)

var (
	lastCommitIDKey = []byte{}
)

func init() {
	f := func(name string) []byte {
		b := make([]byte, 0, 2+len(name))
		b = append(b, []byte{255, MetaType}...)
		b = append(b, name...)
		return b
	}

	lastCommitIDKey = f("last_commit_id")
}

func (l *Ledis) GetLastCommitID() (uint64, error) {
	return Uint64(l.ldb.Get(lastCommitIDKey))
}

func (l *Ledis) GetLastLogID() (uint64, error) {
	if l.log == nil {
		return 0, nil
	}

	return l.log.LastID()
}

func setLastCommitID(t *batch, id uint64) {
	t.Put(lastCommitIDKey, num.Uint64ToBytes(id))
}
