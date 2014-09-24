package ledis

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/go/hack"
	"reflect"
	"strconv"
	"unsafe"
)

var errIntNumber = errors.New("invalid integer")

func Int64(v []byte, err error) (int64, error) {
	if err != nil {
		return 0, err
	} else if v == nil || len(v) == 0 {
		return 0, nil
	} else if len(v) != 8 {
		return 0, errIntNumber
	}

	return int64(binary.LittleEndian.Uint64(v)), nil
}

func Uint64(v []byte, err error) (uint64, error) {
	if err != nil {
		return 0, err
	} else if v == nil || len(v) == 0 {
		return 0, nil
	} else if len(v) != 8 {
		return 0, errIntNumber
	}

	return binary.LittleEndian.Uint64(v), nil
}

func PutInt64(v int64) []byte {
	var b []byte
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pbytes.Data = uintptr(unsafe.Pointer(&v))
	pbytes.Len = 8
	pbytes.Cap = 8
	return b
}

func StrInt64(v []byte, err error) (int64, error) {
	if err != nil {
		return 0, err
	} else if v == nil {
		return 0, nil
	} else {
		return strconv.ParseInt(hack.String(v), 10, 64)
	}
}

func StrUint64(v []byte, err error) (uint64, error) {
	if err != nil {
		return 0, err
	} else if v == nil {
		return 0, nil
	} else {
		return strconv.ParseUint(hack.String(v), 10, 64)
	}
}

func StrInt32(v []byte, err error) (int32, error) {
	if err != nil {
		return 0, err
	} else if v == nil {
		return 0, nil
	} else {
		res, err := strconv.ParseInt(hack.String(v), 10, 32)
		return int32(res), err
	}
}

func StrInt8(v []byte, err error) (int8, error) {
	if err != nil {
		return 0, err
	} else if v == nil {
		return 0, nil
	} else {
		res, err := strconv.ParseInt(hack.String(v), 10, 8)
		return int8(res), err
	}
}

func StrPutInt64(v int64) []byte {
	return strconv.AppendInt(nil, v, 10)
}

func StrPutUint64(v uint64) []byte {
	return strconv.AppendUint(nil, v, 10)
}

func AsyncNotify(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}
