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
)

var (
	Delims = []byte("\r\n")

	NullBulk  = []byte("-1")
	NullArray = []byte("-1")

	PONG          = "PONG"
	OK            = "OK"
	SErrCmdParams = "ERR invalid command param"
	SErrValue     = "ERR value is not an integer or out of range"
	SErrSyntax    = "ERR syntax error"
)
