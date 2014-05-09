package ledis

import (
	"encoding/binary"
	"errors"
	"github.com/siddontang/golib/hack"
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

func PutInt64(v int64) []byte {
	return hack.Int64Slice(v)
}

func PutInt32(v int32) []byte {
	return hack.Int32Slice(v)
}
