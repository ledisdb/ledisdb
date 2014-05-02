package ssdb

import (
	"errors"
)

var (
	ErrEmptyCommand = errors.New("empty command")
	ErrNotFound     = errors.New("command not found")
)

var (
	Delims = []byte("\r\n")

	NullBulk  = []byte("-1")
	NullArray = []byte("-1")
)
