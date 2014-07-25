package store

import (
	"github.com/siddontang/ledisdb/store/driver"
)

type WriteBatch struct {
	driver.IWriteBatch
}
