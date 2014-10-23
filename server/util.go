package server

import (
	"bufio"
	"errors"
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
