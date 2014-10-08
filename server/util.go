package server

import (
	"bufio"
	"errors"
	"github.com/siddontang/go/hack"
	"io"
	"strconv"
)

var (
	errArrayFormat = errors.New("bad array format")
	errBulkFormat  = errors.New("bad bulk string format")
	errLineFormat  = errors.New("bad response line format")
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

func ReadBulkTo(rb *bufio.Reader, w io.Writer) error {
	l, err := ReadLine(rb)
	if err != nil {
		return err
	} else if len(l) == 0 {
		return errBulkFormat
	} else if l[0] == '$' {
		var n int
		//handle resp string
		if n, err = strconv.Atoi(hack.String(l[1:])); err != nil {
			return err
		} else if n == -1 {
			return nil
		} else {
			var nn int64
			if nn, err = io.CopyN(w, rb, int64(n)); err != nil {
				return err
			} else if nn != int64(n) {
				return io.ErrShortWrite
			}

			if l, err = ReadLine(rb); err != nil {
				return err
			} else if len(l) != 0 {
				return errBulkFormat
			}
		}
	} else {
		return errBulkFormat
	}

	return nil
}
