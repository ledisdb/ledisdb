package rpl

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
)

type Log struct {
	ID         uint64
	CreateTime uint32

	Data []byte
}

func NewLog(id uint64, data []byte) *Log {
	l := new(Log)
	l.ID = id
	l.CreateTime = uint32(time.Now().Unix())
	l.Data = data

	return l
}

func (l *Log) HeadSize() int {
	return 16
}

func (l *Log) Size() int {
	return l.HeadSize() + len(l.Data)
}

func (l *Log) Marshal() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, l.HeadSize()+len(l.Data)))
	buf.Reset()

	if err := l.Encode(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (l *Log) Unmarshal(b []byte) error {
	buf := bytes.NewBuffer(b)

	return l.Decode(buf)
}

func (l *Log) Encode(w io.Writer) error {
	buf := make([]byte, l.HeadSize())

	pos := 0
	binary.BigEndian.PutUint64(buf[pos:], l.ID)
	pos += 8

	binary.BigEndian.PutUint32(buf[pos:], l.CreateTime)
	pos += 4

	binary.BigEndian.PutUint32(buf[pos:], uint32(len(l.Data)))

	if n, err := w.Write(buf); err != nil {
		return err
	} else if n != len(buf) {
		return io.ErrShortWrite
	}

	if n, err := w.Write(l.Data); err != nil {
		return err
	} else if n != len(l.Data) {
		return io.ErrShortWrite
	}
	return nil
}

func (l *Log) Decode(r io.Reader) error {
	buf := make([]byte, l.HeadSize())

	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	pos := 0
	l.ID = binary.BigEndian.Uint64(buf[pos:])
	pos += 8

	l.CreateTime = binary.BigEndian.Uint32(buf[pos:])
	pos += 4

	length := binary.BigEndian.Uint32(buf[pos:])

	l.Data = make([]byte, length)
	if _, err := io.ReadFull(r, l.Data); err != nil {
		return err
	}

	return nil
}
