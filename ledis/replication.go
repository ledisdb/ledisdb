package ledis

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/siddontang/go-log/log"
	"io"
	"os"
)

var (
	errInvalidBinLogEvent = errors.New("invalid binglog event")
)

func (l *Ledis) replicateEvent(event []byte) error {
	if len(event) == 0 {
		return errInvalidBinLogEvent
	}

	logType := uint8(event[0])
	switch logType {
	case BinLogTypePut:
		return l.replicatePutEvent(event)
	case BinLogTypeDeletion:
		return l.replicateDeleteEvent(event)
	case BinLogTypeCommand:
		return l.replicateCommandEvent(event)
	default:
		return errInvalidBinLogEvent
	}
}

func (l *Ledis) replicatePutEvent(event []byte) error {
	key, value, err := decodeBinLogPut(event)
	if err != nil {
		return err
	}

	if err = l.ldb.Put(key, value); err != nil {
		return err
	}

	if l.binlog != nil {
		err = l.binlog.Log(event)
	}

	return err
}

func (l *Ledis) replicateDeleteEvent(event []byte) error {
	key, err := decodeBinLogDelete(event)
	if err != nil {
		return err
	}

	if err = l.ldb.Delete(key); err != nil {
		return err
	}

	if l.binlog != nil {
		err = l.binlog.Log(event)
	}

	return err
}

func (l *Ledis) replicateCommandEvent(event []byte) error {
	return errors.New("command event not supported now")
}

func (l *Ledis) RepliateRelayLog(relayLog string, offset int64) (int64, error) {
	f, err := os.Open(relayLog)
	if err != nil {
		return 0, err
	}

	defer f.Close()

	st, _ := f.Stat()
	totalSize := st.Size()

	if _, err = f.Seek(offset, os.SEEK_SET); err != nil {
		return 0, err
	}

	rb := bufio.NewReaderSize(f, 4096)

	var createTime uint32
	var dataLen uint32
	var dataBuf bytes.Buffer

	for {
		if offset+8 > totalSize {
			//event may not sync completely
			return f.Seek(offset, os.SEEK_SET)
		}

		if err = binary.Read(rb, binary.BigEndian, &createTime); err != nil {
			return 0, err
		}

		if err = binary.Read(rb, binary.BigEndian, &dataLen); err != nil {
			return 0, err
		}

		if offset+8+int64(dataLen) > totalSize {
			//event may not sync completely
			return f.Seek(offset, os.SEEK_SET)
		} else {
			if _, err = io.CopyN(&dataBuf, rb, int64(dataLen)); err != nil {
				return 0, err
			}

			l.Lock()
			err = l.replicateEvent(dataBuf.Bytes())
			l.Unlock()
			if err != nil {
				log.Fatal("replication error %s, skip to next", err.Error())
			}

			dataBuf.Reset()

			offset += (8 + int64(dataLen))
		}
	}

	//can not go here???
	log.Error("can not go here")
	return offset, nil
}

func (l *Ledis) WriteRelayLog(data []byte) error {
	return nil
}
