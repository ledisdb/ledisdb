package ledis

import (
	"encoding/binary"
	"testing"
)

func cmpBytes(a []byte, b []byte) bool {
	if len(a) != len(b) {
		println("len diff")
		println(len(a))
		println(len(b))
		return true
	}

	for i, n := range a {
		if n != b[i] {
			println("diff !")
			println(i)
			println(n)
			println(b[i])
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
	testCount(t)
	testOpAndOr(t)
	testOpXor(t)
	testOpNot(t)
	testMSetBit(t)
	testBitExpire(t)
	testBFlush(t)
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

	key := []byte("test_bin_2")

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

func testCount(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	key := []byte("test_bin_count")

	if ori, _ := db.BSetBit(key, 0, 1); ori != 0 {
		t.Error(ori)
	}

	if ori, _ := db.BSetBit(key, 10, 1); ori != 0 {
		t.Error(ori)
	}

	if ori, _ := db.BSetBit(key, 262140, 1); ori != 0 {
		t.Error(ori)
	}

	// count

	if sum, _ := db.BCount(key, 0, -1); sum != 3 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 0, 9); sum != 1 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 0, 10); sum != 2 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 0, 11); sum != 2 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 0, 262139); sum != 2 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 0, 262140); sum != 3 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 0, 262141); sum != 3 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 10, 262140); sum != 2 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 11, 262140); sum != 1 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 11, 262139); sum != 0 {
		t.Error(sum)
	}

	key = []byte("test_bin_count_2")

	db.BSetBit(key, 1, 1)
	db.BSetBit(key, 2, 1)
	db.BSetBit(key, 4, 1)
	db.BSetBit(key, 6, 1)

	if sum, _ := db.BCount(key, 0, -1); sum != 4 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 1, 1); sum != 1 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 0, 7); sum != 4 {
		t.Error(sum)
	}

	if ori, _ := db.BSetBit(key, 8, 1); ori != 0 {
		t.Error(ori)
	}

	if ori, _ := db.BSetBit(key, 11, 1); ori != 0 {
		t.Error(ori)
	}

	if sum, _ := db.BCount(key, 0, -1); sum != 6 {
		t.Error(sum)
	}

	if sum, _ := db.BCount(key, 0, 16); sum != 6 {
		t.Error(sum)
	}
}

func testOpAndOr(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	dstKey := []byte("test_bin_op_and_or")

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

func testOpAnd(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	dstKey := []byte("test_bin_or")

	k0 := []byte("op_or_0")
	k1 := []byte("op_or_01")
	srcKeys := [][]byte{k0, k1}

	db.BSetBit(k0, 0, 1)
	db.BSetBit(k0, 2, 1)

	db.BSetBit(k1, 1, 1)

	if blen, _ := db.BOperation(OPand, dstKey, srcKeys...); blen != 3 {
		t.Fatal(blen)
	}

	if cnt, _ := db.BCount(dstKey, 0, -1); cnt != 1 {
		t.Fatal(1)
	}
}

func testOpXor(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	dstKey := []byte("test_bin_op_xor")

	k0 := []byte("op_xor_00")
	k1 := []byte("op_xor_01")
	srcKeys := [][]byte{k0, k1}

	reqs := make([]BitPair, 4)
	reqs[0] = BitPair{0, 1}
	reqs[1] = BitPair{7, 1}
	reqs[2] = BitPair{int32(segBitSize - 1), 1}
	reqs[3] = BitPair{int32(segBitSize - 8), 1}
	db.BMSetBit(k0, reqs...)

	reqs = make([]BitPair, 2)
	reqs[0] = BitPair{7, 1}
	reqs[1] = BitPair{int32(segBitSize - 8), 1}
	db.BMSetBit(k1, reqs...)

	var stdData []byte
	var data []byte

	// op - xor
	db.BOperation(OPxor, dstKey, srcKeys...)

	stdData = make([]byte, segByteSize)
	stdData[0] = uint8(0x01)
	stdData[segByteSize-1] = uint8(0x80)

	data, _ = db.BGet(dstKey)
	if cmpBytes(data, stdData) {
		t.Fatal(false)
	}
}

func testOpNot(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	//	intputs
	dstKey := []byte("test_bin_op_not")

	k0 := []byte("op_not_0")
	srcKeys := [][]byte{k0}

	db.BSetBit(k0, int32(0), 1)
	db.BSetBit(k0, int32(7), 1)

	pos := segBitSize
	for i := uint32(8); i >= 1; i -= 2 {
		db.BSetBit(k0, int32(pos-i), 1)
	}

	db.BSetBit(k0, int32(3*segBitSize-10), 1)

	//	std
	stdData := make([]byte, segByteSize*3-1)
	for i, _ := range stdData {
		stdData[i] = 255
	}
	stdData[0] = uint8(0x7e)
	stdData[segByteSize-1] = uint8(0xaa)
	stdData[segByteSize*3-2] = uint8(0x3f)

	//	op - not
	db.BOperation(OPnot, dstKey, srcKeys...)

	data, _ := db.BGet(dstKey)
	if cmpBytes(data, stdData) {
		t.Fatal(false)
	}

	k1 := []byte("op_not_2")
	srcKeys = [][]byte{k1}

	db.BSetBit(k1, 0, 1)
	db.BSetBit(k1, 2, 1)
	db.BSetBit(k1, 4, 1)
	db.BSetBit(k1, 6, 1)

	if blen, _ := db.BOperation(OPnot, dstKey, srcKeys...); blen != 7 {
		t.Fatal(blen)
	}

	if cnt, _ := db.BCount(dstKey, 0, -1); cnt != 3 {
		t.Fatal(cnt)
	}
}

func testMSetBit(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	key := []byte("test_mset")

	var datas = make([]BitPair, 8)

	//	1st
	datas[0] = BitPair{1000, 1}
	datas[1] = BitPair{11, 1}
	datas[2] = BitPair{10, 1}
	datas[3] = BitPair{2, 1}
	datas[4] = BitPair{int32(segBitSize - 1), 1}
	datas[5] = BitPair{int32(segBitSize), 1}
	datas[6] = BitPair{int32(segBitSize + 1), 1}
	datas[7] = BitPair{int32(segBitSize) + 10, 0}

	db.BMSetBit(key, datas...)

	if sum, _ := db.BCount(key, 0, -1); sum != 7 {
		t.Error(sum)
	}

	if tail, _ := db.BTail(key); tail != int32(segBitSize+10) {
		t.Error(tail)
	}

	//	2nd
	datas = make([]BitPair, 5)

	datas[0] = BitPair{1000, 0}
	datas[1] = BitPair{int32(segBitSize + 1), 0}
	datas[2] = BitPair{int32(segBitSize * 10), 1}
	datas[3] = BitPair{10, 0}
	datas[4] = BitPair{99, 0}

	db.BMSetBit(key, datas...)

	if sum, _ := db.BCount(key, 0, -1); sum != 7-3+1 {
		t.Error(sum)
	}

	if tail, _ := db.BTail(key); tail != int32(segBitSize*10) {
		t.Error(tail)
	}

	return
}

func testBitExpire(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	key := []byte("test_b_ttl")

	db.BSetBit(key, 0, 1)

	if res, err := db.BExpire(key, 100); res != 1 || err != nil {
		t.Fatal(false)
	}

	if ttl, err := db.BTTL(key); ttl != 100 || err != nil {
		t.Fatal(false)
	}
}

func testBFlush(t *testing.T) {
	db := getTestDB()
	db.FlushAll()

	for i := 0; i < 2000; i++ {
		key := make([]byte, 4)
		binary.LittleEndian.PutUint32(key, uint32(i))
		if _, err := db.BSetBit(key, 1, 1); err != nil {
			t.Fatal(err.Error())
		}
	}

	if v, err := db.BScan(nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 2000 {
		t.Fatal("invalid value ", len(v))
	}

	for i := 0; i < 2000; i++ {
		key := make([]byte, 4)
		binary.LittleEndian.PutUint32(key, uint32(i))
		if v, err := db.BGetBit(key, 1); err != nil {
			t.Fatal(err.Error())
		} else if v != 1 {
			t.Fatal("invalid value ", v)
		}
	}

	if n, err := db.bFlush(); err != nil {
		t.Fatal(err.Error())
	} else if n != 2000 {
		t.Fatal("invalid value ", n)
	}

	if v, err := db.BScan(nil, 3000, true, ""); err != nil {
		t.Fatal(err.Error())
	} else if len(v) != 0 {
		t.Fatal("invalid value length ", len(v))
	}

	for i := 0; i < 2000; i++ {
		key := make([]byte, 4)
		binary.LittleEndian.PutUint32(key, uint32(i))
		if v, err := db.BGet(key); err != nil {
			t.Fatal(err.Error())
		} else if v != nil {

			t.Fatal("invalid value ", v)
		}
	}

}
