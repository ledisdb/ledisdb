package ledis

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/ledisdb/leveldb"
)

const (
	OPand uint8 = iota + 1
	OPor
	OPxor
	OPnot
)

const (
	// byte
	segByteWidth uint32 = 9
	segByteSize  uint32 = 1 << segByteWidth

	// bit
	segBitWidth uint32 = segByteWidth + 3
	segBitSize  uint32 = segByteSize << 3

	maxByteSize uint32 = 8 << 20
	maxSegCount uint32 = maxByteSize / segByteSize

	minSeq uint32 = 0
	maxSeq uint32 = uint32((maxByteSize << 3) - 1)
)

var bitsInByte = [256]int32{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3,
	4, 1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 1, 2, 2, 3, 2, 3,
	3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4,
	5, 5, 6, 1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 2, 3, 3, 4,
	3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4,
	5, 4, 5, 5, 6, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 1, 2,
	2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 2, 3, 3, 4, 3, 4, 4, 5, 3,
	4, 4, 5, 4, 5, 5, 6, 2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 2, 3, 3, 4, 3, 4, 4,
	5, 3, 4, 4, 5, 4, 5, 5, 6, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6,
	6, 7, 3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 4, 5, 5, 6, 5,
	6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8}

var fillSegment []byte = func() []byte {
	data := make([]byte, segByteSize, segByteSize)
	for i := uint32(0); i < segByteSize; i++ {
		data[i] = 0xff
	}
	return data
}()

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
	pos++

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
	seq = uint32(binary.BigEndian.Uint32(bkey[4+keyLen:]))
	return
}

func (db *DB) bCapByteSize(seq uint32, off uint32) uint32 {
	var offByteSize uint32 = (off >> 3) + 1
	if offByteSize > segByteSize {
		offByteSize = segByteSize
	}

	return seq<<segByteWidth + offByteSize
}

func (db *DB) bParseOffset(key []byte, offset int32) (seq uint32, off uint32, err error) {
	if offset < 0 {
		if tailSeq, tailOff, e := db.bGetMeta(key); e != nil {
			err = e
			return
		} else if tailSeq >= 0 {
			offset += int32(uint32(tailSeq)<<segBitWidth | uint32(tailOff))
			if offset < 0 {
				err = errOffset
				return
			}
		}
	}

	off = uint32(offset)

	seq = off >> segBitWidth
	off &= (segBitSize - 1)
	return
}

func (db *DB) bGetMeta(key []byte) (tailSeq int32, tailOff int32, err error) {
	var v []byte

	mk := db.bEncodeMetaKey(key)
	v, err = db.db.Get(mk)
	if err != nil {
		return
	}

	if v != nil {
		tailSeq = int32(binary.LittleEndian.Uint32(v[0:4]))
		tailOff = int32(binary.LittleEndian.Uint32(v[4:8]))
	} else {
		tailSeq = -1
		tailOff = -1
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
	var ts, to int32
	if ts, to, err = db.bGetMeta(key); err != nil {
		return
	} else {
		tailSeq = uint32(MaxInt32(ts, 0))
		tailOff = uint32(MaxInt32(to, 0))
	}

	if seq > tailSeq || (seq == tailSeq && off > tailOff) {
		db.bSetMeta(t, key, seq, off)
		tailSeq = seq
		tailOff = off
	}
	return
}

func (db *DB) bDelete(t *tx, key []byte) (drop int64) {
	mk := db.bEncodeMetaKey(key)
	t.Delete(mk)

	minKey := db.bEncodeBinKey(key, minSeq)
	maxKey := db.bEncodeBinKey(key, maxSeq)
	it := db.db.RangeIterator(minKey, maxKey, leveldb.RangeClose)
	for ; it.Valid(); it.Next() {
		t.Delete(it.Key())
		drop = 1
	}
	it.Close()

	return drop
}

func (db *DB) BGet(key []byte) (data []byte, err error) {
	if err = checkKeySize(key); err != nil {
		return
	}

	var ts, to int32
	if ts, to, err = db.bGetMeta(key); err != nil || ts < 0 {
		return
	}

	var tailSeq, tailOff = uint32(ts), uint32(to)
	var capByteSize uint32 = db.bCapByteSize(tailSeq, tailOff)
	data = make([]byte, capByteSize, capByteSize)

	minKey := db.bEncodeBinKey(key, minSeq)
	maxKey := db.bEncodeBinKey(key, tailSeq)
	it := db.db.RangeIterator(minKey, maxKey, leveldb.RangeClose)

	var seq, s, e uint32
	for ; it.Valid(); it.Next() {
		if _, seq, err = db.bDecodeBinKey(it.Key()); err != nil {
			data = nil
			break
		}

		s = seq << segByteWidth
		e = MinUInt32(s+segByteSize, capByteSize)
		copy(data[s:e], it.Value())
	}
	it.Close()

	return
}

func (db *DB) BDelete(key []byte) (drop int64, err error) {
	if err = checkKeySize(key); err != nil {
		return
	}

	t := db.binTx
	t.Lock()
	defer t.Unlock()

	drop = db.bDelete(t, key)
	db.rmExpire(t, bExpType, key)

	err = t.Commit()
	return
}

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
		segment = make([]byte, segByteSize, segByteSize) // can be optimize ?
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

	it := db.db.RangeIterator(skey, ekey, leveldb.RangeClose)
	for ; it.Valid(); it.Next() {
		segment = it.Value()
		for _, bit := range segment {
			cnt += bitsInByte[bit]
		}
	}
	it.Close()

	return
}

func (db *DB) BTail(key []byte) (int32, error) {
	// effective length of data, the highest bit-pos set in history
	tailSeq, tailOff, err := db.bGetMeta(key)
	if err != nil {
		return 0, err
	}

	tail := int32(-1)
	if tailSeq >= 0 {
		tail = int32(uint32(tailSeq)<<segBitWidth | uint32(tailOff))
	}

	return tail, nil
}

func (db *DB) bSegAnd(a []byte, b []byte, res *[]byte) {
	if a == nil || b == nil {
		*res = nil
		return
	}

	data := *res
	if data == nil {
		data = make([]byte, segByteSize, segByteSize)
		*res = data
	}

	for i := uint32(0); i < segByteSize; i++ {
		data[i] = a[i] & b[i]
	}
	return
}

func (db *DB) bSegOr(a []byte, b []byte, res *[]byte) {
	if a == nil || b == nil {
		if a == nil && b == nil {
			*res = nil
		} else if a == nil {
			*res = b
		} else {
			*res = a
		}
		return
	}

	data := *res
	if data == nil {
		data = make([]byte, segByteSize, segByteSize)
		*res = data
	}

	for i := uint32(0); i < segByteSize; i++ {
		data[i] = a[i] | b[i]
	}
	return
}

func (db *DB) bIterator(key []byte) *leveldb.RangeLimitIterator {
	sk := db.bEncodeBinKey(key, minSeq)
	ek := db.bEncodeBinKey(key, maxSeq)
	return db.db.RangeIterator(sk, ek, leveldb.RangeClose)
}

func (db *DB) BOperation(op uint8, dstkey []byte, srckeys ...[]byte) (blen int32, err error) {
	//	return :
	//		The size of the string stored in the destination key,
	//		that is equal to the size of the longest input string.
	var exeOp func([]byte, []byte, *[]byte)
	switch op {
	case OPand:
		exeOp = db.bSegAnd
	case OPor:
		exeOp = db.bSegOr
	default:
		return
	}

	if dstkey == nil || srckeys == nil || len(srckeys) == 0 {
		return
	}

	t := db.binTx
	t.Lock()
	defer t.Unlock()

	var seq, off uint32
	var segments = make([][]byte, maxSegCount) // todo : init count also can be optimize while 'and' / 'or'

	// init - meta info
	var maxDstSeq, maxDstOff uint32
	var nowSeq, nowOff int32

	if nowSeq, nowOff, err = db.bGetMeta(srckeys[0]); err != nil { // todo : if key not exists ....
		return
	} else if nowSeq < 0 {
		return // incorrect ...
	} else {
		maxDstSeq = uint32(nowSeq)
		maxDstOff = uint32(nowOff)
	}

	// init - data
	it := db.bIterator(srckeys[0])
	for ; it.Valid(); it.Next() {
		if _, seq, err = db.bDecodeBinKey(it.Key()); err != nil {
			// to do ...
			it.Close()
			return
		}
		segments[seq] = it.Value()
	}
	it.Close()

	//	operation with following keys
	var keyNum int = len(srckeys)
	for i := 1; i < keyNum; i++ {
		if nowSeq, nowOff, err = db.bGetMeta(srckeys[i]); err != nil {
			return
		}

		if nowSeq < 0 {
			continue
		} else {
			seq = uint32(nowSeq)
			off = uint32(nowOff)
			if seq > maxDstSeq || (seq == maxDstSeq && off > maxDstOff) {
				maxDstSeq = seq
				maxDstOff = off
			}
		}

		it = db.bIterator(srckeys[i])
		segIdx := uint32(0)

		for end := false; !end; it.Next() {
			end = !it.Valid()
			if !end {
				if _, seq, err = db.bDecodeBinKey(it.Key()); err != nil {
					// to do ...
					it.Close()
					return
				}
			} else {
				seq = maxSegCount
			}

			// todo :
			// 		operation 'and' can be optimize here :
			//		if seq > max_segments_idx, this loop can be break,
			//		which can avoid cost from Key() and decode key

			if op == OPand || op == OPor {
				for ; segIdx < seq; segIdx++ {
					if segments[segIdx] != nil {
						exeOp(segments[segIdx], nil, &segments[segIdx])
					}
				}
			} // else {...}

			if !end {
				exeOp(segments[seq], it.Value(), &segments[segIdx])
				segIdx++
			}
		}
		it.Close()
	}

	// clear the old data in case
	db.bDelete(t, dstkey)
	db.rmExpire(t, bExpType, dstkey)

	//	set data and meta
	if op == OPand || op == OPor {
		// for i := maxDstSeq; i >= 0; i-- {
		//
		// }
		db.bSetMeta(t, dstkey, maxDstSeq, maxDstOff)
	}

	var bk []byte
	for seq, seg := range segments {
		if seg != nil {
			//	todo:
			//		here can be optimize, like 'updateBinKeySeq',
			//		avoid too many make mem
			bk = db.bEncodeBinKey(dstkey, uint32(seq))
			t.Put(bk, seg)
		}
	}

	err = t.Commit()
	return
}

// func (db *DB) BExpire(key []byte, duration int64) (int64, error) {

// }

// func (db *DB) BExpireAt(key []byte, when int64) (int64, error) {

// }

// func (db *DB) BTTL(key []byte) (int64, error) {

// }

// func (db *DB) BPersist(key []byte) (int64, error) {

// }

// func (db *DB) BScan(key []byte, count int, inclusive bool) ([]KVPair, error) {

// }

// func (db *DB) bFlush() (drop int64, err error) {

// }
