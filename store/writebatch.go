package store

import (
	"github.com/siddontang/ledisdb/store/driver"
)

type WriteBatch interface {
	driver.IWriteBatch
}
