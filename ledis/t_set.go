package ledis

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/ledisdb/store"
	"time"
)

var errSetKey = errors.New("invalid set key")
var errSSizeKey = errors.New("invalid ssize key")

const (
	setStartSep byte = ':'
	setStopSep  byte = setStartSep + 1
)

func checkSetKMSize(key []byte, member []byte) error {
	if len(key) > MaxKeySize || len(key) == 0 {
		return errKeySize
	} else if len(member) > MaxSetMemberSize || len(member) == 0 {
		return errSetMemberSize
	}
	return nil
}

func (db *DB) sEncodeSizeKey(key []byte) []byte {
	buf := make([]byte, len(key)+2)

	buf[0] = db.index
	buf[1] = SSizeType

	copy(buf[2:], key)
	return buf
}

func (db *DB) sDecodeSizeKey(ek []byte) ([]byte, error) {
	if len(ek) < 2 || ek[0] != db.index || ek[1] != SSizeType {
		return nil, errSSizeKey
	}

	return ek[2:], nil
}

func (db *DB) sEncodeSetKey(key []byte, member []byte) []byte {
	buf := make([]byte, len(key)+len(member)+1+1+2+1)

	pos := 0
	buf[pos] = db.index
	pos++
	buf[pos] = SetType
	pos++

	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2

	copy(buf[pos:], key)
	pos += len(key)

	buf[pos] = setStartSep
	pos++
	copy(buf[pos:], member)

	return buf
}

func (db *DB) sDecodeSetKey(ek []byte) ([]byte, []byte, error) {
	if len(ek) < 5 || ek[0] != db.index || ek[1] != SetType {
		return nil, nil, errSetKey
	}

	pos := 2
	keyLen := int(binary.BigEndian.Uint16(ek[pos:]))
	pos += 2

	if keyLen+5 > len(ek) {
		return nil, nil, errSetKey
	}

	key := ek[pos : pos+keyLen]
	pos += keyLen

	if ek[pos] != hashStartSep {
		return nil, nil, errSetKey
	}

	pos++
	member := ek[pos:]
	return key, member, nil
}

func (db *DB) sEncodeStartKey(key []byte) []byte {
	return db.sEncodeSetKey(key, nil)
}

func (db *DB) sEncodeStopKey(key []byte) []byte {
	k := db.sEncodeSetKey(key, nil)

	k[len(k)-1] = setStopSep

	return k
}

// func (db *DB) sFlush() {

// }

func (db *DB) sDelete(t *tx, key []byte) int64 {
	sk := db.sEncodeSizeKey(key)
	start := db.sEncodeStartKey(key)
	stop := db.sEncodeStopKey(key)

	var num int64 = 0
	it := db.db.RangeLimitIterator(start, stop, store.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		num++
	}

	it.Close()

	t.Delete(sk)
	return num
}

func (db *DB) sIncrSize(key []byte, delta int64) (int64, error) {
	t := db.setTx
	sk := db.sEncodeSizeKey(key)

	var err error
	var size int64 = 0
	if size, err = Int64(db.db.Get(sk)); err != nil {
		return 0, err
	} else {
		size += delta
		if size <= 0 {
			size = 0
			t.Delete(sk)
			db.rmExpire(t, SetType, key)
		} else {
			t.Put(sk, PutInt64(size))
		}
	}

	return size, nil
}

func (db *DB) sExpireAt(key []byte, when int64) (int64, error) {
	t := db.setTx
	t.Lock()
	defer t.Unlock()

	if scnt, err := db.SCard(key); err != nil || scnt == 0 {
		return 0, err
	} else {
		db.expireAt(t, SetType, key, when)
		if err := t.Commit(); err != nil {
			return 0, err
		}

	}

	return 1, nil
}

func (db *DB) sSetItem(key []byte, member []byte) (int64, error) {
	t := db.setTx
	ek := db.sEncodeSetKey(key, member)

	var n int64 = 1
	if v, _ := db.db.Get(ek); v != nil {
		n = 0
	} else {
		if _, err := db.sIncrSize(key, 1); err != nil {
			return 0, err
		}
	}

	t.Put(ek, nil)
	return n, nil
}

func (db *DB) SAdd(key []byte, args ...[]byte) (int64, error) {
	t := db.setTx
	t.Lock()
	defer t.Unlock()

	var err error
	var ek []byte
	var num int64 = 0
	for i := 0; i < len(args); i++ {
		if err := checkSetKMSize(key, args[i]); err != nil {
			return 0, err
		}

		ek = db.sEncodeSetKey(key, args[i])

		if v, err := db.db.Get(ek); err != nil {
			return 0, err
		} else if v == nil {
			num++
		}

		t.Put(ek, nil)
	}

	if _, err = db.sIncrSize(key, num); err != nil {
		return 0, err
	}

	err = t.Commit()
	return num, err

}

func (db *DB) SCard(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	sk := db.sEncodeSizeKey(key)

	return Int64(db.db.Get(sk))
}

// TODO
// func (db *DB) sDiffGeneric()

// func (db *DB) SDiff() {

// }

// func (db *DB) SDiffStore() {

// }

// func (db *DB) SInter() {

// }

// func (db *DB) SInterStore() {

// }

func (db *DB) SIsMember(key []byte, member []byte) (int64, error) {
	ek := db.sEncodeSetKey(key, member)

	var n int64 = 1
	if v, err := db.db.Get(ek); err != nil {
		return 0, err
	} else if v == nil {
		n = 0
	}
	return n, nil
}

func (db *DB) SMembers(key []byte) ([][]byte, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	start := db.sEncodeStartKey(key)
	stop := db.sEncodeStopKey(key)

	v := make([][]byte, 0, 16)

	it := db.db.RangeLimitIterator(start, stop, store.RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		_, m, err := db.sDecodeSetKey(it.Key())
		if err != nil {
			return nil, err
		}

		v = append(v, m)
	}

	it.Close()

	return v, nil
}

func (db *DB) SRem(key []byte, args ...[]byte) (int64, error) {
	t := db.setTx
	t.Lock()
	defer t.Unlock()

	var ek []byte
	var v []byte
	var err error

	it := db.db.NewIterator()
	defer it.Close()

	var num int64 = 0
	for i := 0; i < len(args); i++ {
		if err := checkSetKMSize(key, args[i]); err != nil {
			return 0, err
		}

		ek = db.sEncodeSetKey(key, args[i])

		v = it.RawFind(ek)
		if v == nil {
			continue
		} else {
			num++
			t.Delete(ek)
		}
	}

	if _, err = db.sIncrSize(key, -num); err != nil {
		return 0, err
	}

	err = t.Commit()
	return num, err

}

// TODO
// func (db *DB) sUnionGeneric(keys ...[]byte) ([][]byte, error) {

// 	for _, key := range keys {
// 		if err := checkKeySize(key); err != nil {
// 			return nil, err
// 		}

// 	}
// }

// func (db *DB) SUnion(keys ...[]byte) ([][]byte, error) {

// 	if v, err := db.sUnionGeneric(keys); err != nil {
// 		return nil, err
// 	} else if v == nil {
// 		return nil, nil
// 	} else {
// 		return v, nil
// 	}

// }

// func (db *DB) SUnionStore(dstkey []byte, keys []byte) (int64, error) {
// 	if err := checkKeySize(dstkey); err != nil {
// 		return 0, err
// 	}

// }

func (db *DB) SClear(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	t := db.setTx
	t.Lock()
	defer t.Unlock()

	num := db.sDelete(t, key)
	db.rmExpire(t, SetType, key)

	err := t.Commit()
	return num, err
}

func (db *DB) SMclear(keys ...[]byte) (int64, error) {
	t := db.setTx
	t.Lock()
	defer t.Unlock()

	for _, key := range keys {
		if err := checkKeySize(key); err != nil {
			return 0, err
		}

		db.sDelete(t, key)
		db.rmExpire(t, SetType, key)
	}

	err := t.Commit()
	return int64(len(keys)), err
}

func (db *DB) SExpire(key []byte, duration int64) (int64, error) {
	if duration <= 0 {
		return 0, errExpireValue
	}

	return db.sExpireAt(key, time.Now().Unix()+duration)

}

func (db *DB) SExpireAt(key []byte, when int64) (int64, error) {
	if when <= time.Now().Unix() {
		return 0, errExpireValue
	}

	return db.sExpireAt(key, when)

}

func (db *DB) STTL(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return -1, err
	}

	return db.ttl(SetType, key)
}

func (db *DB) SPersist(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	t := db.setTx
	t.Lock()
	defer t.Unlock()

	n, err := db.rmExpire(t, SetType, key)
	if err != nil {
		return 0, err
	}
	err = t.Commit()
	return n, err
}
