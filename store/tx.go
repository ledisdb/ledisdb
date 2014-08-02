package store

import (
	"github.com/siddontang/ledisdb/store/driver"
)

type Tx interface {
	driver.Tx
}
