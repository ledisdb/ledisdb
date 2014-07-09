package ledis

import (
	"errors"
)

const (
	noneType    byte = 0
	kvType      byte = 1
	hashType    byte = 2
	hSizeType   byte = 3
	listType    byte = 4
	lMetaType   byte = 5
	zsetType    byte = 6
	zSizeType   byte = 7
	zScoreType  byte = 8
	binType     byte = 9
	binMetaType byte = 10

	maxDataType byte = 100

	expTimeType byte = 101
	expMetaType byte = 102
)

const (
	defaultScanCount int = 10
)

var (
	errKeySize        = errors.New("invalid key size")
	errValueSize      = errors.New("invalid value size")
	errHashFieldSize  = errors.New("invalid hash field size")
	errZSetMemberSize = errors.New("invalid zset member size")
	errExpireValue    = errors.New("invalid expire value")
)

const (
	//we don't support too many databases
	MaxDBNumber uint8 = 16

	//max key size
	MaxKeySize int = 1024

	//max hash field size
	MaxHashFieldSize int = 1024

	//max zset member size
	MaxZSetMemberSize int = 1024

	//max value size
	MaxValueSize int = 10 * 1024 * 1024
)

var (
	ErrScoreMiss = errors.New("zset score miss")
)

const (
	BinLogTypeDeletion uint8 = 0x0
	BinLogTypePut      uint8 = 0x1
	BinLogTypeCommand  uint8 = 0x2
)
