package server

import (
	"errors"

	"github.com/siddontang/ledisdb/ledis"
)

var (
	ErrEmptyCommand          = errors.New("empty command")
	ErrNotFound              = errors.New("command not found")
	ErrNotAuthenticated      = errors.New("not authenticated")
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrCmdParams             = errors.New("invalid command param")
	ErrValue                 = errors.New("value is not an integer or out of range")
	ErrSyntax                = errors.New("syntax error")
	ErrOffset                = errors.New("offset bit is not an natural number")
	ErrBool                  = errors.New("value is not 0 or 1")
)

var (
	Delims = []byte("\r\n")

	NullBulk  = []byte("-1")
	NullArray = []byte("-1")

	PONG  = "PONG"
	OK    = "OK"
	NOKEY = "NOKEY"
)

const (
	KV   ledis.DataType = ledis.KV
	LIST                = ledis.LIST
	HASH                = ledis.HASH
	SET                 = ledis.SET
	ZSET                = ledis.ZSET
)

const (
	KVName   = ledis.KVName
	ListName = ledis.ListName
	HashName = ledis.HashName
	SetName  = ledis.SetName
	ZSetName = ledis.ZSetName
)

const (
	GB uint64 = 1024 * 1024 * 1024
	MB uint64 = 1024 * 1024
	KB uint64 = 1024
)

var TypeNames = []string{KVName, ListName, HashName, SetName, ZSetName}
