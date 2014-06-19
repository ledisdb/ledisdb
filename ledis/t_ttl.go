package ledis

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/ledisdb/leveldb"
	"time"
)

var mapExpMetaType = map[byte]byte{
	kvExpType: kvExpMetaType,
	lExpType:  lExpMetaType,
	hExpType:  hExpMetaType,
	zExpType:  zExpMetaType}

type retireCallback func(*tx, []byte) int64

type elimination struct {
	db         *DB
	exp2Tx     map[byte]*tx
	exp2Retire map[byte]retireCallback
}

var errExpType = errors.New("invalid expire type")

func (db *DB) expEncodeTimeKey(expType byte, key []byte, when int64) []byte {
	// format : db[8] / expType[8] / when[64] / key[...]
	buf := make([]byte, len(key)+10)

	buf[0] = db.index
	buf[1] = expType
	pos := 2

	binary.BigEndian.PutUint64(buf[pos:], uint64(when))
	pos += 8

	copy(buf[pos:], key)

	return buf
}

func (db *DB) expEncodeMetaKey(expType byte, key []byte) []byte {
	// format : db[8] / expType[8] / key[...]
	buf := make([]byte, len(key)+2)

	buf[0] = db.index
	buf[1] = expType
	pos := 2

	copy(buf[pos:], key)

	return buf
}

// usage : separate out the original key
func (db *DB) expDecodeMetaKey(mk []byte) []byte {
	if len(mk) <= 2 {
		//	check db ? check type ?
		return nil
	}

	return mk[2:]
}

func (db *DB) expire(t *tx, expType byte, key []byte, duration int64) {
	db.expireAt(t, expType, key, time.Now().Unix()+duration)
}

func (db *DB) expireAt(t *tx, expType byte, key []byte, when int64) {
	mk := db.expEncodeMetaKey(expType+1, key)
	tk := db.expEncodeTimeKey(expType, key, when)

	t.Put(tk, mk)
	t.Put(mk, PutInt64(when))
}

func (db *DB) ttl(expType byte, key []byte) (t int64, err error) {
	mk := db.expEncodeMetaKey(expType+1, key)

	if t, err = Int64(db.db.Get(mk)); err != nil || t == 0 {
		t = -1
	} else {
		t -= time.Now().Unix()
		if t <= 0 {
			t = -1
		}
		// if t == -1 : to remove ????
	}

	return t, err
}

func (db *DB) rmExpire(t *tx, expType byte, key []byte) {
	mk := db.expEncodeMetaKey(expType+1, key)
	if v, err := db.db.Get(mk); err != nil || v == nil {
		return
	} else if when, err2 := Int64(v, nil); err2 != nil {
		return
	} else {
		tk := db.expEncodeTimeKey(expType, key, when)
		t.Delete(mk)
		t.Delete(tk)
	}
}

func (db *DB) expFlush(t *tx, expType byte) (err error) {
	expMetaType, ok := mapExpMetaType[expType]
	if !ok {
		return errExpType
	}

	drop := 0

	minKey := make([]byte, 2)
	minKey[0] = db.index
	minKey[1] = expType

	maxKey := make([]byte, 2)
	maxKey[0] = db.index
	maxKey[1] = expMetaType + 1

	it := db.db.RangeLimitIterator(minKey, maxKey, leveldb.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		drop++
		if drop&1023 == 0 {
			if err = t.Commit(); err != nil {
				return
			}
		}
	}
	it.Close()

	err = t.Commit()
	return
}

//////////////////////////////////////////////////////////
//
//////////////////////////////////////////////////////////

func newEliminator(db *DB) *elimination {
	eli := new(elimination)
	eli.db = db
	eli.exp2Tx = make(map[byte]*tx)
	eli.exp2Retire = make(map[byte]retireCallback)
	return eli
}

func (eli *elimination) regRetireContext(expType byte, t *tx, onRetire retireCallback) {
	eli.exp2Tx[expType] = t
	eli.exp2Retire[expType] = onRetire
}

//	call by outside ... (from *db to another *db)
func (eli *elimination) active() {
	now := time.Now().Unix()
	db := eli.db
	dbGet := db.db.Get
	expKeys := make([][]byte, 0, 1024)
	expTypes := [...]byte{kvExpType, lExpType, hExpType, zExpType}

	for _, et := range expTypes {
		//	search those keys' which expire till the moment
		minKey := db.expEncodeTimeKey(et, nil, 0)
		maxKey := db.expEncodeTimeKey(et, nil, now+1)
		expKeys = expKeys[0:0]

		t, _ := eli.exp2Tx[et]
		onRetire, _ := eli.exp2Retire[et]
		if t == nil || onRetire == nil {
			// todo : log error
			continue
		}

		it := db.db.RangeLimitIterator(minKey, maxKey, leveldb.RangeROpen, 0, -1)
		for it.Valid() {
			for i := 1; i < 512 && it.Valid(); i++ {
				expKeys = append(expKeys, it.Key(), it.Value())
				it.Next()
			}

			var cnt int = len(expKeys)
			if cnt == 0 {
				continue
			}

			t.Lock()
			var mk, ek, k []byte
			for i := 0; i < cnt; i += 2 {
				ek, mk = expKeys[i], expKeys[i+1]
				if exp, err := Int64(dbGet(mk)); err == nil {
					// check expire again
					if exp > now {
						continue
					}

					// delete keys
					k = db.expDecodeMetaKey(mk)
					onRetire(t, k)
					t.Delete(ek)
					t.Delete(mk)
				}
			}
			t.Commit()
			t.Unlock()
		} // end : it
		it.Close()
	} // end : expType

	return
}
