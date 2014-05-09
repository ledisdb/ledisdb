package ledis

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

const (
	KV_TYPE byte = iota + 1
	HASH_TYPE
	HSIZE_TYPE
	LIST_TYPE
	LMETA_TYPE
	ZSET_TYPE
	ZSIZE_TYPE
	ZSCORE_TYPE
)
