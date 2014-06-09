package server

import (
	"bufio"
	"errors"
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"strconv"
)

var (
	errArrayFormat = errors.New("bad array format")
	errBulkFormat  = errors.New("bad bulk string format")
	errLineFormat  = errors.New("bad response line format")
)

func readLine(rb *bufio.Reader) ([]byte, error) {
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

func readBulkTo(rb *bufio.Reader, w io.Writer) error {
	l, err := readLine(rb)
	if len(l) == 0 {
		return errArrayFormat
	} else if l[0] == '$' {
		var n int
		//handle resp string
		if n, err = strconv.Atoi(ledis.String(l[1:])); err != nil {
			return err
		} else if n == -1 {
			return nil
		} else {
			if _, err = io.CopyN(w, rb, int64(n)); err != nil {
				return err
			}

			if l, err = readLine(rb); err != nil {
				return err
			} else if len(l) != 0 {
				return errBulkFormat
			}
		}
	} else {
		return errArrayFormat
	}

	return nil
}
