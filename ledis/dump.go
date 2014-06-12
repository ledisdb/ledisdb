package ledis

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/siddontang/go-leveldb/leveldb"
	"io"
	"os"
)

//dump format
// fileIndex(bigendian int64)|filePos(bigendian int64)
// |keylen(bigendian int32)|key|valuelen(bigendian int32)|value......

type MasterInfo struct {
	LogFileIndex int64
	LogPos       int64
}

func (m *MasterInfo) WriteTo(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, m.LogFileIndex); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, m.LogPos); err != nil {
		return err
	}
	return nil
}

func (m *MasterInfo) ReadFrom(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &m.LogFileIndex)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &m.LogPos)
	if err != nil {
		return err
	}

	return nil
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
	var m *MasterInfo = new(MasterInfo)
	if l.binlog == nil {
		sp = l.ldb.NewSnapshot()
	} else {
		l.Lock()
		sp = l.ldb.NewSnapshot()
		m.LogFileIndex = l.binlog.LogFileIndex()
		m.LogPos = l.binlog.LogFilePos()
		l.Unlock()
	}

	var err error

	wb := bufio.NewWriterSize(w, 4096)
	if err = m.WriteTo(wb); err != nil {
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

func (l *Ledis) LoadDumpFile(path string) (*MasterInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return l.LoadDump(f)
}

func (l *Ledis) LoadDump(r io.Reader) (*MasterInfo, error) {
	l.Lock()
	defer l.Unlock()

	info := new(MasterInfo)

	rb := bufio.NewReaderSize(r, 4096)

	err := info.ReadFrom(rb)
	if err != nil {
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

	return info, nil
}
