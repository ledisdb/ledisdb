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
	//testSimple(t)
	testSimpleII(t)
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

	if tail, _ := db.BTail(key); tail != uint32(50) {
		t.Error(tail)
	}
}

func testSimpleII(t *testing.T) {
	db := getTestDB()
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

	if tail, _ := db.BTail(key); tail != uint32(pos) {
		t.Error(tail)
	}

	data, _ := db.BGet(key)
	stdData := newBytes(pos + 1)
	stdData[pos/8] = uint8(1 << (uint(pos) % 8))

	if cmpBytes(data, stdData) {
		t.Error(len(data))
	}
}
