package ledis

import (
	"encoding/binary"
	"errors"
)

var errSetKey = errors.New("invalid set key")
var errSSizeKey = errors.New("invalid ssize key")

const (
	setStartSep byte = ':'
	setStopSep  byte = setStartSep + 1
)

func checkSetKeyMember(key []byte, member []byte) error {
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
