package ledis

import (
	"encoding/binary"
	"errors"
)

var (
	errBinLogDeleteType  = errors.New("invalid bin log delete type")
	errBinLogPutType     = errors.New("invalid bin log put type")
	errBinLogCommandType = errors.New("invalid bin log command type")
)

func encodeBinLogDelete(key []byte) []byte {
	buf := make([]byte, 1+len(key))
	buf[0] = BinLogTypeDeletion
	copy(buf[1:], key)
	return buf
}

func decodeBinLogDelete(sz []byte) ([]byte, error) {
	if len(sz) < 1 || sz[0] != BinLogTypeDeletion {
		return nil, errBinLogDeleteType
	}

	return sz[1:], nil
}

func encodeBinLogPut(key []byte, value []byte) []byte {
	buf := make([]byte, 3+len(key)+len(value))
	buf[0] = BinLogTypePut
	pos := 1
	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2
	copy(buf[pos:], key)
	pos += len(key)
	copy(buf[pos:], value)

	return buf
}

func decodeBinLogPut(sz []byte) ([]byte, []byte, error) {
	if len(sz) < 3 || sz[0] != BinLogTypePut {
		return nil, nil, errBinLogPutType
	}

	keyLen := int(binary.BigEndian.Uint16(sz[1:]))
	if 3+keyLen > len(sz) {
		return nil, nil, errBinLogPutType
	}

	return sz[3 : 3+keyLen], sz[3+keyLen:], nil
}

func encodeBinLogCommand(commandType uint8, args ...[]byte) []byte {
	//to do
	return nil
}

func decodeBinLogCommand(sz []byte) (uint8, [][]byte, error) {
	return 0, nil, errBinLogCommandType
}
