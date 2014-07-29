package server

import (
	"errors"
)

var (
	ErrEmptyCommand = errors.New("empty command")
	ErrNotFound     = errors.New("command not found")
	ErrCmdParams    = errors.New("invalid command param")
	ErrValue        = errors.New("value is not an integer or out of range")
	ErrSyntax       = errors.New("syntax error")
	ErrOffset       = errors.New("offset bit is not an natural number")
	ErrBool         = errors.New("value is not 0 or 1")
)

var (
	Delims = []byte("\r\n")

	NullBulk  = []byte("-1")
	NullArray = []byte("-1")

	PONG = "PONG"
	OK   = "OK"
)
