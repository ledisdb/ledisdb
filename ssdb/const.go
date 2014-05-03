package ssdb

import (
	"errors"
)

var (
	ErrEmptyCommand = errors.New("empty command")
	ErrNotFound     = errors.New("command not found")
	ErrCmdParams    = errors.New("invalid command param")
)

var (
	Delims = []byte("\r\n")

	NullBulk  = []byte("-1")
	NullArray = []byte("-1")

	PONG = "PONG"
	OK   = "OK"
)
