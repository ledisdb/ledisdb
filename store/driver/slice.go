package driver

// #include <stdlib.h>
import "C"

import (
	"reflect"
	"unsafe"
)

type ISlice interface {
	Data() []byte
	Size() int
	Free()
}

type CSlice struct {
	data unsafe.Pointer
	size int
}

func NewCSlice(p unsafe.Pointer, n int) *CSlice {
	return &CSlice{p, n}
}

func (s *CSlice) Data() []byte {
	var value []byte

	sH := (*reflect.SliceHeader)(unsafe.Pointer(&value))
	sH.Cap = int(s.size)
	sH.Len = int(s.size)
	sH.Data = uintptr(s.data)

	return value
}

func (s *CSlice) Size() int {
	return int(s.size)
}

func (s *CSlice) Free() {
	if s.data != nil {
		C.free(s.data)
		s.data = nil
	}
}

type GoSlice []byte

func (s GoSlice) Data() []byte {
	return []byte(s)
}

func (s GoSlice) Size() int {
	return len(s)
}

func (s GoSlice) Free() {

}
