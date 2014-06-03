package ledis

import (
	"time"
)

func (db *DB) FlushAll() (drop int64, err error) {
	all := [...](func() (int64, error)){
		db.Flush,
		db.LFlush,
		db.HFlush,
		db.ZFlush}

	for _, flush := range all {
		if n, e := flush(); e != nil {
			err = e
			return
		} else {
			drop += n
		}
	}

	return
}

func (db *DB) activeExpireCycle() {
	eliminator := newEliminator(db)
	eliminator.regRetireContext(kvExpType, db.kvTx, db.delete)

	go func() {
		for {
			eliminator.active()
			time.Sleep(1 * time.Second)
		}
	}()
}
