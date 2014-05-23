package ledis

import (
	"errors"
)

const (
	kvType byte = iota + 1
	hashType
	hSizeType
	listType
	lMetaType
	zsetType
	zSizeType
	zScoreType
)

const (
	defaultScanCount int = 20
)

const (
	//we don't support too many databases
	MaxDBNumber uint8 = 16

	//max key size
	MaxKeySize int = 1<<16 - 1

	//max hash field size
	MaxHashFieldSize int = 1<<16 - 1

	//max zset member size
	MaxZSetMemberSize int = 1<<16 - 1
)

var (
	ErrKeySize        = errors.New("invalid key size")
	ErrHashFieldSize  = errors.New("invalid hash field size")
	ErrZSetMemberSize = errors.New("invalid zset member size")
)
