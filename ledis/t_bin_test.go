package ledis

import (
	"testing"
)

func cmpBytes(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return true
	}

	for i, n := range a {
		if n != b[i] {
			return true
		}
	}
	return false
}

func newBytes(bitLen int32) []byte {
	bytes := bitLen / 8
	if bitLen%8 > 0 {
		bytes++
	}

	return make([]byte, bytes, bytes)
}

func TestBinary(t *testing.T) {
	testSimple(t)
	testSimpleII(t)
	testOp(t)
}

func testSimple(t *testing.T) {
	db := getTestDB()

	key := []byte("test_bin")

	if v, _ := db.BGetBit(key, 100); v != 0 {
		t.Error(v)
	}

	if ori, _ := db.BSetBit(key, 50, 1); ori != 0 {
		t.Error(ori)
	}

	if v, _ := db.BGetBit(key, 50); v != 1 {
		t.Error(v)
	}

	if ori, _ := db.BSetBit(key, 50, 0); ori != 1 {
		t.Error(ori)
	}

	if v, _ := db.BGetBit(key, 50); v != 0 {
		t.Error(v)
	}

	db.BSetBit(key, 7, 1)
	db.BSetBit(key, 8, 1)
	db.BSetBit(key, 9, 1)
	db.BSetBit(key, 10, 1)

	if sum, _ := db.BCount(key, 0, -1); sum != 4 {
		t.Error(sum)
	}

	data, _ := db.BGet(key)
	if cmpBytes(data, []byte{0x80, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Error(data)
	}

	if tail, _ := db.BTail(key); tail != int32(50) {
		t.Error(tail)
	}
}

func testSimpleII(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	key := []byte("test_bin")

	pos := int32(1234567)
	if ori, _ := db.BSetBit(key, pos, 1); ori != 0 {
		t.Error(ori)
	}

	if v, _ := db.BGetBit(key, pos); v != 1 {
		t.Error(v)
	}

	if v, _ := db.BGetBit(key, pos-1); v != 0 {
		t.Error(v)
	}

	if v, _ := db.BGetBit(key, pos+1); v != 0 {
		t.Error(v)
	}

	if tail, _ := db.BTail(key); tail != pos {
		t.Error(tail)
	}

	data, _ := db.BGet(key)
	stdData := newBytes(pos + 1)
	stdData[pos/8] = uint8(1 << (uint(pos) % 8))

	if cmpBytes(data, stdData) {
		t.Error(len(data))
	}

	if drop, _ := db.BDelete(key); drop != 1 {
		t.Error(false)
	}

	if data, _ := db.BGet(key); data != nil {
		t.Error(data)
	}
}

func testOp(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	dstKey := []byte("test_bin_op")

	k0 := []byte("op_0")
	k1 := []byte("op_01")
	k2 := []byte("op_10")
	k3 := []byte("op_11")
	srcKeys := [][]byte{k2, k0, k1, k3}

	/*
		<k0>
			<seg> - <high> ... <low>
			0 - [10000000] ... [00000001]
			1 - nil
			2 - [00000000] ... [11111111] ... [00000000]
			3 - [01010101] ... [10000001] [10101010]
			4 - [10000000] ... [00000000]
			5 - [00000000] ... [00000011] [00000001]
			...
	*/
	// (k0 - seg:0)
	db.BSetBit(k0, int32(0), 1)
	db.BSetBit(k0, int32(segBitSize-1), 1)
	// (k0 - seg:2)
	pos := segBitSize*2 + segBitSize/2
	for i := uint32(0); i < 8; i++ {
		db.BSetBit(k0, int32(pos+i), 1)
	}
	// (k0 - seg:3)
	pos = segBitSize * 3
	db.BSetBit(k0, int32(pos+8), 1)
	db.BSetBit(k0, int32(pos+15), 1)
	for i := uint32(1); i < 8; i += 2 {
		db.BSetBit(k0, int32(pos+i), 1)
	}
	pos = segBitSize*4 - 8
	for i := uint32(0); i < 8; i += 2 {
		db.BSetBit(k0, int32(pos+i), 1)
	}
	// (k0 - seg:4)
	db.BSetBit(k0, int32(segBitSize*5-1), 1)
	// (k0 - seg:5)
	db.BSetBit(k0, int32(segBitSize*5), 1)
	db.BSetBit(k0, int32(segBitSize*5+8), 1)
	db.BSetBit(k0, int32(segBitSize*5+9), 1)

	/*
		<k1>
			0 - nil
			1 - [00000001] ... [10000000]
			2 - nil
			3 - [10101010] ... [10000001] [01010101]
			...
	*/
	// (k1 - seg:1)
	db.BSetBit(k1, int32(segBitSize+7), 1)
	db.BSetBit(k1, int32(segBitSize*2-8), 1)
	// (k1 - seg:3)
	pos = segBitSize * 3
	db.BSetBit(k1, int32(pos+8), 1)
	db.BSetBit(k1, int32(pos+15), 1)
	for i := uint32(0); i < 8; i += 2 {
		db.BSetBit(k0, int32(pos+i), 1)
	}
	pos = segBitSize*4 - 8
	for i := uint32(1); i < 8; i += 2 {
		db.BSetBit(k0, int32(pos+i), 1)
	}

	var stdData []byte
	var data []byte
	var tmpKeys [][]byte

	//	op - or
	db.BOperation(OPor, dstKey, srcKeys...)

	stdData = make([]byte, 5*segByteSize+2)
	stdData[0] = uint8(0x01)
	stdData[segByteSize-1] = uint8(0x80)
	stdData[segByteSize] = uint8(0x80)
	stdData[segByteSize*2-1] = uint8(0x01)
	stdData[segByteSize*2+segByteSize/2] = uint8(0xff)
	stdData[segByteSize*3] = uint8(0xff)
	stdData[segByteSize*3+1] = uint8(0x81)
	stdData[segByteSize*4-1] = uint8(0xff)
	stdData[segByteSize*5-1] = uint8(0x80)
	stdData[segByteSize*5] = uint8(0x01)
	stdData[segByteSize*5+1] = uint8(0x03)

	data, _ = db.BGet(dstKey)
	if cmpBytes(data, stdData) {
		t.Fatal(false)
	}

	tmpKeys = [][]byte{k0, dstKey, k1}
	db.BOperation(OPor, dstKey, tmpKeys...)

	data, _ = db.BGet(dstKey)
	if cmpBytes(data, stdData) {
		t.Fatal(false)
	}

	//	op - and
	db.BOperation(OPand, dstKey, srcKeys...)

	stdData = make([]byte, 5*segByteSize+2)
	stdData[segByteSize*3+1] = uint8(0x81)

	data, _ = db.BGet(dstKey)
	if cmpBytes(data, stdData) {
		t.Fatal(false)
	}

	tmpKeys = [][]byte{k0, dstKey, k1}
	db.BOperation(OPand, dstKey, tmpKeys...)

	data, _ = db.BGet(dstKey)
	if cmpBytes(data, stdData) {
		t.Fatal(false)
	}

}
