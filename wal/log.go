package wal

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Log struct {
	ID         uint64
	CreateTime uint32
	// 0 for no compression
	// 1 for snappy compression
	Compression uint8
	Data        []byte
}

func (l *Log) Marshal() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 17+len(l.Data)))
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
	length := uint32(17)
	buf := make([]byte, length)

	pos := 0
	binary.BigEndian.PutUint64(buf[pos:], l.ID)
	pos += 8

	binary.BigEndian.PutUint32(buf[pos:], l.CreateTime)
	pos += 4

	buf[pos] = l.Compression
	pos++

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
	length := uint32(17)
	buf := make([]byte, length)

	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	pos := 0
	l.ID = binary.BigEndian.Uint64(buf[pos:])
	pos += 8

	l.CreateTime = binary.BigEndian.Uint32(buf[pos:])
	pos += 4

	l.Compression = buf[pos]
	pos++

	length = binary.BigEndian.Uint32(buf[pos:])

	l.Data = make([]byte, length)
	if _, err := io.ReadFull(r, l.Data); err != nil {
		return err
	}

	return nil
}
