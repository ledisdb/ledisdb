package server

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/siddontang/go/arena"
	"io"
)

var (
	errLineFormat = errors.New("bad response line format")
)

func ReadLine(rb *bufio.Reader) ([]byte, error) {
	p, err := rb.ReadSlice('\n')

	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, errLineFormat
	}

	return p[:i], nil
}

func readBytes(br *bufio.Reader, a *arena.Arena) (bytes []byte, err error) {
	size, err := readLong(br)
	if err != nil {
		return nil, err
	}
	if size == -1 {
		return nil, nil
	}
	if size < 0 {
		return nil, errors.New("Invalid size: " + fmt.Sprint("%d", size))
	}

	buf := a.Make(int(size) + 2)
	if _, err = io.ReadFull(br, buf); err != nil {
		return nil, err
	}

	if buf[len(buf)-2] != '\r' && buf[len(buf)-1] != '\n' {
		return nil, errors.New("bad bulk string format")
	}

	bytes = buf[0 : len(buf)-2]

	return
}

func readLong(in *bufio.Reader) (result int64, err error) {
	read, err := in.ReadByte()
	if err != nil {
		return -1, err
	}
	var sign int
	if read == '-' {
		read, err = in.ReadByte()
		if err != nil {
			return -1, err
		}
		sign = -1
	} else {
		sign = 1
	}
	var number int64
	for number = 0; err == nil; read, err = in.ReadByte() {
		if read == '\r' {
			read, err = in.ReadByte()
			if err != nil {
				return -1, err
			}
			if read == '\n' {
				return number * int64(sign), nil
			} else {
				return -1, errors.New("Bad line ending")
			}
		}
		value := read - '0'
		if value >= 0 && value < 10 {
			number *= 10
			number += int64(value)
		} else {
			return -1, errors.New("Invalid digit")
		}
	}
	return -1, err
}

func ReadRequest(in *bufio.Reader, a *arena.Arena) ([][]byte, error) {
	code, err := in.ReadByte()
	if err != nil {
		return nil, err
	}

	if code != '*' {
		return nil, errReadRequest
	}

	var nparams int64
	if nparams, err = readLong(in); err != nil {
		return nil, err
	} else if nparams <= 0 {
		return nil, errReadRequest
	}

	req := make([][]byte, nparams)
	for i := range req {
		if code, err = in.ReadByte(); err != nil {
			return nil, err
		} else if code != '$' {
			return nil, errReadRequest
		}

		if req[i], err = readBytes(in, a); err != nil {
			return nil, err
		}
	}

	return req, nil
}

func lowerSlice(buf []byte) []byte {
	for i, r := range buf {
		if 'A' <= r && r <= 'Z' {
			r += 'a' - 'A'
		}

		buf[i] = r
	}
	return buf
}
