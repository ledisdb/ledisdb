package rpl

import (
	"os"
)

type table struct {
	f *os.File

	first uint64
	last  uint64
}
