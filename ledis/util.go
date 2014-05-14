package ledis

import (
	"encoding/binary"
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

var errIntNumber = errors.New("invalid integer")

// no copy to change slice to string
// use your own risk
func String(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}

// no copy to change string to slice
// use your own risk
func Slice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}

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
		return strconv.ParseInt(String(v), 10, 64)
	}
}

func StrPutInt64(v int64) []byte {
	return Slice(strconv.FormatInt(v, 10))
}
