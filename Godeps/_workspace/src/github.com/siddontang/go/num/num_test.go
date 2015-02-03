package num

import (
	"testing"
)

func testMin(t *testing.T, v1 interface{}, v2 interface{}, v interface{}) {
	var c interface{}
	switch i1 := v1.(type) {
	case int:
		c = MinInt(i1, v2.(int))
	case int8:
		c = MinInt8(i1, v2.(int8))
	case int16:
		c = MinInt16(i1, v2.(int16))
	case int32:
		c = MinInt32(i1, v2.(int32))
	case int64:
		c = MinInt64(i1, v2.(int64))
	case uint:
		c = MinUint(i1, v2.(uint))
	case uint8:
		c = MinUint8(i1, v2.(uint8))
	case uint16:
		c = MinUint16(i1, v2.(uint16))
	case uint32:
		c = MinUint32(i1, v2.(uint32))
	case uint64:
		c = MinUint64(i1, v2.(uint64))
	default:
		t.Fatalf("invalid type %T", t)
	}

	if c != v {
		t.Fatalf("invalid %v(%T) != %v(%T)", c, c, v, v)
	}
}

func TestMin(t *testing.T) {
	testMin(t, int(1), int(2), int(1))
	testMin(t, int(1), int(1), int(1))

	testMin(t, int8(1), int8(2), int8(1))
	testMin(t, int8(1), int8(1), int8(1))

	testMin(t, int16(1), int16(2), int16(1))
	testMin(t, int16(1), int16(1), int16(1))

	testMin(t, int32(1), int32(2), int32(1))
	testMin(t, int32(1), int32(1), int32(1))

	testMin(t, int64(1), int64(2), int64(1))
	testMin(t, int64(1), int64(1), int64(1))

	testMin(t, uint(1), uint(2), uint(1))
	testMin(t, uint(1), uint(1), uint(1))

	testMin(t, uint8(1), uint8(2), uint8(1))
	testMin(t, uint8(1), uint8(1), uint8(1))

	testMin(t, uint16(1), uint16(2), uint16(1))
	testMin(t, uint16(1), uint16(1), uint16(1))

	testMin(t, uint32(1), uint32(2), uint32(1))
	testMin(t, uint32(1), uint32(1), uint32(1))

	testMin(t, uint64(1), uint64(2), uint64(1))
	testMin(t, uint64(1), uint64(1), uint64(1))
}

func testMax(t *testing.T, v1 interface{}, v2 interface{}, v interface{}) {
	var c interface{}
	switch i1 := v1.(type) {
	case int:
		c = MaxInt(i1, v2.(int))
	case int8:
		c = MaxInt8(i1, v2.(int8))
	case int16:
		c = MaxInt16(i1, v2.(int16))
	case int32:
		c = MaxInt32(i1, v2.(int32))
	case int64:
		c = MaxInt64(i1, v2.(int64))
	case uint:
		c = MaxUint(i1, v2.(uint))
	case uint8:
		c = MaxUint8(i1, v2.(uint8))
	case uint16:
		c = MaxUint16(i1, v2.(uint16))
	case uint32:
		c = MaxUint32(i1, v2.(uint32))
	case uint64:
		c = MaxUint64(i1, v2.(uint64))
	default:
		t.Fatalf("invalid type %T", t)
	}

	if c != v {
		t.Fatalf("invalid %v(%T) != %v(%T)", c, c, v, v)
	}
}

func TestMax(t *testing.T) {
	testMax(t, int(1), int(2), int(2))
	testMax(t, int(1), int(1), int(1))

	testMax(t, int8(1), int8(2), int8(2))
	testMax(t, int8(1), int8(1), int8(1))

	testMax(t, int16(1), int16(2), int16(2))
	testMax(t, int16(1), int16(1), int16(1))

	testMax(t, int32(1), int32(2), int32(2))
	testMax(t, int32(1), int32(1), int32(1))

	testMax(t, int64(1), int64(2), int64(2))
	testMax(t, int64(1), int64(1), int64(1))

	testMax(t, uint(1), uint(2), uint(2))
	testMax(t, uint(1), uint(1), uint(1))

	testMax(t, uint8(1), uint8(2), uint8(2))
	testMax(t, uint8(1), uint8(1), uint8(1))

	testMax(t, uint16(1), uint16(2), uint16(2))
	testMax(t, uint16(1), uint16(1), uint16(1))

	testMax(t, uint32(1), uint32(2), uint32(2))
	testMax(t, uint32(1), uint32(1), uint32(1))

	testMax(t, uint64(1), uint64(2), uint64(2))
	testMax(t, uint64(1), uint64(1), uint64(1))
}

func TestBytes(t *testing.T) {
	if BytesToUint64(Uint64ToBytes(1)) != 1 {
		t.Fatal("convert fail")
	}

	if BytesToUint32(Uint32ToBytes(1)) != 1 {
		t.Fatal("convert fail")
	}

	if BytesToUint16(Uint16ToBytes(1)) != 1 {
		t.Fatal("convert fail")
	}

	if BytesToInt64(Int64ToBytes(-1)) != -1 {
		t.Fatal("convert fail")
	}

	if BytesToInt32(Int32ToBytes(-1)) != -1 {
		t.Fatal("convert fail")
	}

	if BytesToInt16(Int16ToBytes(-1)) != -1 {
		t.Fatal("convert fail")
	}
}

func TestStr(t *testing.T) {
	if v, err := ParseUint64(FormatUint64(1)); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Fatal(v)
	}

	if v, err := ParseUint32(FormatUint32(1)); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Fatal(v)
	}

	if v, err := ParseUint16(FormatUint16(1)); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Fatal(v)
	}

	if v, err := ParseUint8(FormatUint8(1)); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Fatal(v)
	}

	if v, err := ParseInt64(FormatInt64(-1)); err != nil {
		t.Fatal(err)
	} else if v != -1 {
		t.Fatal(v)
	}

	if v, err := ParseInt32(FormatInt32(-1)); err != nil {
		t.Fatal(err)
	} else if v != -1 {
		t.Fatal(v)
	}

	if v, err := ParseInt16(FormatInt16(-1)); err != nil {
		t.Fatal(err)
	} else if v != -1 {
		t.Fatal(v)
	}

	if v, err := ParseInt8(FormatInt8(-1)); err != nil {
		t.Fatal(err)
	} else if v != -1 {
		t.Fatal(v)
	}
}
