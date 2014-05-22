package ledis

func (db *DB) Flush() (drop int64, err error) {
	all := [...](func() (int64, error)){
		db.KvFlush,
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
