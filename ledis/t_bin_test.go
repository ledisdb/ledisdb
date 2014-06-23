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

func TestBinary(t *testing.T) {
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
}
