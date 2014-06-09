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
	errInvalidBinLogFile  = errors.New("invalid binlog file")
)

func (l *Ledis) ReplicateEvent(event []byte) error {
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

func (l *Ledis) ReplicateFromReader(rb io.Reader) error {
	var createTime uint32
	var dataLen uint32
	var dataBuf bytes.Buffer
	var err error

	for {
		if err = binary.Read(rb, binary.BigEndian, &createTime); err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		if err = binary.Read(rb, binary.BigEndian, &dataLen); err != nil {
			return err
		}

		if _, err = io.CopyN(&dataBuf, rb, int64(dataLen)); err != nil {
			return err
		}

		err = l.ReplicateEvent(dataBuf.Bytes())
		if err != nil {
			log.Fatal("replication error %s, skip to next", err.Error())
		}

		dataBuf.Reset()
	}

	return nil
}

func (l *Ledis) ReplicateFromData(data []byte) error {
	rb := bytes.NewReader(data)

	l.Lock()
	err := l.ReplicateFromReader(rb)
	l.Unlock()

	return err
}

func (l *Ledis) ReplicateFromBinLog(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	rb := bufio.NewReaderSize(f, 4096)

	err = l.ReplicateFromReader(rb)

	f.Close()

	return err
}

const maxSyncEvents = 16

//events data format
// nextfileIndex(bigendian int64)|nextfilePos(bigendian int64)|binlogevents
func (l *Ledis) ReadEventsTo(index int64, offset int64, w io.Writer) (info *MasterInfo, err error) {
	info = new(MasterInfo)

	if l.binlog == nil {
		//binlog not supported
		info.LogFileIndex = 0
		return
	}

	info.LogFileIndex = index
	info.LogPos = offset

	filePath := l.binlog.FormatLogFilePath(index)

	var f *os.File
	f, err = os.Open(filePath)
	if err != nil && !os.IsNotExist(err) {
		return
	} else if os.IsNotExist(err) {
		l.Lock()
		lastIndex := l.binlog.LogFileIndex()
		if index == lastIndex {
			//no binlog at all
			l.Unlock()
			return
		}
		l.Unlock()

		//slave binlog info had lost
		info.LogFileIndex = -1
	}

	defer f.Close()

	if _, err = f.Seek(offset, os.SEEK_SET); err != nil {
		//may be invliad seek offset
		return
	}

	var lastCreateTime uint32 = 0
	var createTime uint32
	var dataLen uint32

	var n int = 0

	for {
		if err = binary.Read(f, binary.BigEndian, &createTime); err != nil {
			if err == io.EOF {
				//we will try to use next binlog
				info.LogFileIndex = index + 1
				info.LogPos = 0

				return
			} else {
				return
			}
		}

		n++
		if lastCreateTime == 0 {
			lastCreateTime = createTime
		} else if lastCreateTime != createTime {
			return
		} else if n > maxSyncEvents {
			return
		}

		if err = binary.Read(f, binary.BigEndian, &dataLen); err != nil {
			return
		}

		if err = binary.Write(w, binary.BigEndian, createTime); err != nil {
			return
		}

		if err = binary.Write(w, binary.BigEndian, dataLen); err != nil {
			return
		}

		if _, err = io.CopyN(w, f, int64(dataLen)); err != nil {
			return
		}

		info.LogPos = info.LogPos + 8 + int64(dataLen)
	}

	return
}
