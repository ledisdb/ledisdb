package ledis

import (
	"encoding/binary"
)

func encodeBinLogDelete(key []byte) []byte {
	buf := make([]byte, 3+len(key))
	buf[0] = BinLogTypeDeletion
	pos := 1
	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2
	copy(buf[pos:], key)
	return buf
}

func encodeBinLogPut(key []byte, value []byte) []byte {
	buf := make([]byte, 7+len(key)+len(value))
	buf[0] = BinLogTypePut
	pos := 1
	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2
	copy(buf[pos:], key)
	pos += len(key)
	binary.BigEndian.PutUint32(buf[pos:], uint32(len(value)))
	pos += 4
	copy(buf[pos:], value)
	return buf
}

func encodeBinLogCommand(commandType uint8, args []byte) []byte {
	//to do
	return nil
}
