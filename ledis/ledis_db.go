package ledis

func (db *DB) FlushAll() (drop int64, err error) {
	all := [...](func() (int64, error)){
		db.flush,
		db.lFlush,
		db.hFlush,
		db.zFlush}

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

func (db *DB) newEliminator() *elimination {
	eliminator := newEliminator(db)
	eliminator.regRetireContext(kvExpType, db.kvTx, db.delete)
	eliminator.regRetireContext(lExpType, db.listTx, db.lDelete)
	eliminator.regRetireContext(hExpType, db.hashTx, db.hDelete)
	eliminator.regRetireContext(zExpType, db.zsetTx, db.zDelete)

	return eliminator
}
