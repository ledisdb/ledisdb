package rpl

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Log struct {
	ID          uint64
	CreateTime  uint32
	Compression uint8

	Data []byte
}

func (l *Log) HeadSize() int {
	return 17
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
	if err := binary.Write(w, binary.BigEndian, l.ID); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, l.CreateTime); err != nil {
		return err
	}

	if _, err := w.Write([]byte{l.Compression}); err != nil {
		return err
	}

	dataLen := uint32(len(l.Data))
	if err := binary.Write(w, binary.BigEndian, dataLen); err != nil {
		return err
	}

	if n, err := w.Write(l.Data); err != nil {
		return err
	} else if n != len(l.Data) {
		return io.ErrShortWrite
	}
	return nil
}

func (l *Log) Decode(r io.Reader) error {
	length, err := l.DecodeHead(r)
	if err != nil {
		return err
	}

	l.Data = l.Data[0:0]

	if cap(l.Data) >= int(length) {
		l.Data = l.Data[0:length]
	} else {
		l.Data = make([]byte, length)
	}
	if _, err := io.ReadFull(r, l.Data); err != nil {
		return err
	}

	return nil
}

func (l *Log) DecodeHead(r io.Reader) (uint32, error) {
	buf := make([]byte, l.HeadSize())

	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}

	pos := 0
	l.ID = binary.BigEndian.Uint64(buf[pos:])
	pos += 8

	l.CreateTime = binary.BigEndian.Uint32(buf[pos:])
	pos += 4

	l.Compression = uint8(buf[pos])
	pos++

	length := binary.BigEndian.Uint32(buf[pos:])

	return length, nil
}
