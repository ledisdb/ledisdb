package ledis

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/ledisdb/leveldb"
)

const (
	OPand byte = iota + 1
	OPor
	OPxor
	OPnot
)

const (
	segWidth uint32 = 9
	segSize  uint32 = uint32(1 << segWidth) // byte

	minSeq uint32 = 0
	maxSeq uint32 = uint32((1 << 31) - 1)
)

var bitsInByte = [256]int32{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4, 1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 4, 5, 5, 6, 5, 6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8}

var errBinKey = errors.New("invalid bin key")
var errOffset = errors.New("invalid offset")

func getBit(sz []byte, offset uint32) uint8 {
	index := offset >> 3
	if index >= uint32(len(sz)) {
		return 0 // error("overflow")
	}

	offset -= index << 3
	return sz[index] >> offset & 1
}

func setBit(sz []byte, offset uint32, val uint8) bool {
	if val != 1 && val != 0 {
		return false // error("invalid val")
	}

	index := offset >> 3
	if index >= uint32(len(sz)) {
		return false // error("overflow")
	}

	offset -= index << 3
	if sz[index]>>offset&1 != val {
		sz[index] ^= (1 << offset)
	}
	return true
}

func (db *DB) bEncodeMetaKey(key []byte) []byte {
	mk := make([]byte, len(key)+2)
	mk[0] = db.index
	mk[1] = binMetaType

	copy(mk, key)
	return mk
}

func (db *DB) bEncodeBinKey(key []byte, seq uint32) []byte {
	bk := make([]byte, len(key)+8)

	pos := 0
	bk[pos] = db.index
	pos++
	bk[pos] = binType

	binary.BigEndian.PutUint16(bk[pos:], uint16(len(key)))
	pos += 2

	copy(bk[pos:], key)
	pos += len(key)

	binary.BigEndian.PutUint32(bk[pos:], seq)

	return bk
}

func (db *DB) bDecodeBinKey(bkey []byte) (key []byte, seq uint32, err error) {
	if len(bkey) < 8 || bkey[0] != db.index {
		err = errBinKey
		return
	}

	keyLen := binary.BigEndian.Uint16(bkey[2:4])
	if int(keyLen+8) != len(bkey) {
		err = errBinKey
		return
	}

	key = bkey[4 : 4+keyLen]
	seq = uint32(binary.BigEndian.Uint16(bkey[4+keyLen:]))
	return
}

func (db *DB) bParseOffset(key []byte, offset int32) (seq uint32, off uint32, err error) {
	if offset < 0 {
		if tailSeq, tailOff, e := db.bGetMeta(key); e != nil {
			err = e
			return
		} else {
			offset += int32(tailSeq<<segWidth | tailOff)
			if offset < 0 {
				err = errOffset
				return
			}
		}
	}

	off = uint32(offset)

	seq = off >> segWidth
	off &= (segSize - 1)
	return
}

func (db *DB) bGetMeta(key []byte) (tailSeq uint32, tailOff uint32, err error) {
	var v []byte

	mk := db.bEncodeMetaKey(key)
	v, err = db.db.Get(mk)
	if err != nil {
		return
	}

	if v != nil {
		tailSeq = binary.LittleEndian.Uint32(v[0:4])
		tailOff = binary.LittleEndian.Uint32(v[4:8])
	}
	return
}

func (db *DB) bSetMeta(t *tx, key []byte, tailSeq uint32, tailOff uint32) {
	ek := db.bEncodeMetaKey(key)

	//	todo ..
	//	if size == 0 // headSeq == tailSeq
	//		t.Delete(ek)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], tailSeq)
	binary.LittleEndian.PutUint32(buf[4:8], tailOff)

	t.Put(ek, buf)
	return
}

func (db *DB) bUpdateMeta(t *tx, key []byte, seq uint32, off uint32) (tailSeq uint32, tailOff uint32, err error) {
	if tailSeq, tailOff, err = db.bGetMeta(key); err != nil {
		return
	}

	if seq > tailSeq || (seq == tailSeq && off > tailOff) {
		db.bSetMeta(t, key, seq, off)
		tailSeq = seq
		tailOff = off
	}
	return
}

// func (db *DB) bDelete(key []byte) int64 {
// 	return 0
// }

func (db *DB) BGet(key []byte) (data []byte, err error) {
	if err = checkKeySize(key); err != nil {
		return
	}

	var tailSeq, tailOff uint32
	if tailSeq, tailOff, err = db.bGetMeta(key); err != nil {
		return
	}

	var offByteLen uint32 = tailOff >> 3
	if tailOff&7 > 0 {
		offByteLen++
	}

	var dataCap uint32 = tailSeq<<3 + offByteLen
	data = make([]byte, dataCap)

	minKey := db.bEncodeBinKey(key, minSeq)
	maxKey := db.bEncodeBinKey(key, tailSeq)
	it := db.db.RangeLimitIterator(minKey, maxKey, leveldb.RangeClose, 0, -1)
	for pos, end := uint32(0), uint32(0); it.Valid(); it.Next() {
		end = pos + segSize
		if end >= offByteLen {
			end = offByteLen
		}

		copy(data[pos:end], it.Value())
		pos = end
	}
	it.Close()

	return
}

// func (db *DB) BDelete(key []byte) (int8, error) {

// }

func (db *DB) getSegment(key []byte, seq uint32) ([]byte, []byte, error) {
	bk := db.bEncodeBinKey(key, seq)
	segment, err := db.db.Get(bk)
	if err != nil {
		return bk, nil, err
	}
	return bk, segment, nil
}

func (db *DB) allocateSegment(key []byte, seq uint32) ([]byte, []byte, error) {
	bk, segment, err := db.getSegment(key, seq)
	if err == nil && segment == nil {
		segment = make([]byte, segSize, segSize) // can be optimize ?
	}
	return bk, segment, err
}

func (db *DB) BSetBit(key []byte, offset int32, val uint8) (ori uint8, err error) {
	if err = checkKeySize(key); err != nil {
		return
	}

	//	todo : check offset
	var seq, off uint32
	if seq, off, err = db.bParseOffset(key, offset); err != nil {
		return 0, err
	}

	var bk, segment []byte
	if bk, segment, err = db.allocateSegment(key, seq); err != nil {
		return 0, err
	}

	if segment != nil {
		ori = getBit(segment, off)
		setBit(segment, off, val)

		t := db.binTx
		t.Lock()
		t.Put(bk, segment)
		if _, _, e := db.bUpdateMeta(t, key, seq, off); e != nil {
			err = e
			return
		}
		err = t.Commit()
		t.Unlock()
	}

	return
}

func (db *DB) BGetBit(key []byte, offset int32) (uint8, error) {
	if seq, off, err := db.bParseOffset(key, offset); err != nil {
		return 0, err
	} else {
		_, segment, err := db.getSegment(key, seq)
		if err != nil {
			return 0, err
		}

		if segment == nil {
			return 0, nil
		} else {
			return getBit(segment, off), nil
		}
	}
}

// func (db *DB) BGetRange(key []byte, start int32, end int32) ([]byte, error) {
// 	section := make([]byte)

// 	return
// }

func (db *DB) BCount(key []byte, start int32, end int32) (cnt int32, err error) {
	var sseq uint32
	if sseq, _, err = db.bParseOffset(key, start); err != nil {
		return
	}

	var eseq uint32
	if eseq, _, err = db.bParseOffset(key, end); err != nil {
		return
	}

	var segment []byte
	skey := db.bEncodeBinKey(key, sseq)
	ekey := db.bEncodeBinKey(key, eseq)

	it := db.db.RangeLimitIterator(skey, ekey, leveldb.RangeClose, 0, -1)
	for ; it.Valid(); it.Next() {
		segment = it.Value()
		for _, bit := range segment {
			cnt += bitsInByte[bit]
		}
	}
	it.Close()

	return
}

// func (db *DB) BLen(key []) (uint32, error) {

// }

// func (db *DB) BOperation(op byte, dstkey []byte, srckeys ...[]byte) (int32, error) {
// 	//	return :
// 	//		The size of the string stored in the destination key,
// 	//		that is equal to the size of the longest input string.
// 	if op < OPand || op > OPnot {
// 		return
// 	}

// }

// func (db *DB) BExpire(key []byte, duration int64) (int64, error) {

// }

// func (db *DB) BExpireAt(key []byte, when int64) (int64, error) {

// }

// func (db *DB) BTTL(key []byte) (int64, error) {

// }

// func (db *DB) BScan(key []byte, count int, inclusive bool) ([]KVPair, error) {

// }

// func (db *DB) bFlush() (drop int64, err error) {

// }
