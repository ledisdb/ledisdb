package rpl

import (
	"fmt"
	"os"
	"path"
)

type table struct {
	baseName string

	index int64

	readonly bool
	df       *os.File
	mf       *os.File

	first uint64
	last  uint64
}

func newReadTable(base string, index int64) (*table, error) {
	t := new(table)

	t.baseName = path.Join(base, fmt.Sprintf("%08d", index))
	t.index = index

	return t, nil
}
