package ledis

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/siddontang/go-leveldb/leveldb"
	"io"
	"os"
)

//dump format
// head len(bigendian int32)|head(json format)
// |keylen(bigendian int32)|key|valuelen(bigendian int32)|value......

type DumpHead struct {
	LogFile string `json:"log_file"`
	LogPos  int64  `json:"log_pos"`
}

func (l *Ledis) DumpFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return l.Dump(f)
}

func (l *Ledis) Dump(w io.Writer) error {
	var sp *leveldb.Snapshot
	var logFileName string
	var logPos int64
	if l.binlog == nil {
		sp = l.ldb.NewSnapshot()
	} else {
		l.Lock()
		sp = l.ldb.NewSnapshot()
		logFileName = l.binlog.LogFileName()
		logPos = l.binlog.LogFilePos()
		l.Unlock()
	}

	var head = DumpHead{
		LogFile: logFileName,
		LogPos:  logPos,
	}

	data, err := json.Marshal(&head)
	if err != nil {
		return err
	}

	wb := bufio.NewWriterSize(w, 4096)
	if err = binary.Write(wb, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	if _, err = wb.Write(data); err != nil {
		return err
	}

	it := sp.Iterator(nil, nil, leveldb.RangeClose, 0, -1)
	var key []byte
	var value []byte
	for ; it.Valid(); it.Next() {
		key = it.Key()
		value = it.Value()

		if err = binary.Write(wb, binary.BigEndian, uint16(len(key))); err != nil {
			return err
		}

		if _, err = wb.Write(key); err != nil {
			return err
		}

		if err = binary.Write(wb, binary.BigEndian, uint32(len(value))); err != nil {
			return err
		}

		if _, err = wb.Write(value); err != nil {
			return err
		}
	}

	if err = wb.Flush(); err != nil {
		return err
	}

	return nil
}

func (l *Ledis) LoadDumpFile(path string) (*DumpHead, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return l.LoadDump(f)
}

func (l *Ledis) LoadDump(r io.Reader) (*DumpHead, error) {
	l.Lock()
	defer l.Unlock()

	rb := bufio.NewReaderSize(r, 4096)

	var headLen uint32
	err := binary.Read(rb, binary.BigEndian, &headLen)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, headLen)
	if _, err = io.ReadFull(rb, buf); err != nil {
		return nil, err
	}

	var head DumpHead
	if err = json.Unmarshal(buf, &head); err != nil {
		return nil, err
	}

	var keyLen uint16
	var valueLen uint32

	var keyBuf bytes.Buffer
	var valueBuf bytes.Buffer
	for {
		if err = binary.Read(rb, binary.BigEndian, &keyLen); err != nil && err != io.EOF {
			return nil, err
		} else if err == io.EOF {
			break
		}

		if _, err = io.CopyN(&keyBuf, rb, int64(keyLen)); err != nil {
			return nil, err
		}

		if err = binary.Read(rb, binary.BigEndian, &valueLen); err != nil {
			return nil, err
		}

		if _, err = io.CopyN(&valueBuf, rb, int64(valueLen)); err != nil {
			return nil, err
		}

		if err = l.ldb.Put(keyBuf.Bytes(), valueBuf.Bytes()); err != nil {
			return nil, err
		}

		if l.binlog != nil {
			err = l.binlog.Log(encodeBinLogPut(keyBuf.Bytes(), valueBuf.Bytes()))
		}

		keyBuf.Reset()
		valueBuf.Reset()
	}

	return &head, nil
}
