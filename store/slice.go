package store

import (
	"github.com/ledisdb/ledisdb/store/driver"
)

type Slice interface {
	driver.ISlice
}
