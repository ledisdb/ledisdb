package mdb

import (
	"testing"
	"syscall"
)

func TestErrno(t *testing.T) {
	zeroerr := errno(0)
	if zeroerr != nil {
		t.Errorf("errno(0) != nil: %#v", zeroerr)
	}
	syserr := _errno(int(syscall.EINVAL))
	if syserr != syscall.EINVAL { // fails if syserr is Errno(syscall.EINVAL)
		t.Errorf("errno(syscall.EINVAL) != syscall.EINVAL: %#v", syserr)
	}
	mdberr := _errno(int(KeyExist))
	if mdberr != KeyExist { // fails if syserr is Errno(syscall.EINVAL)
		t.Errorf("errno(KeyExist) != KeyExist: %#v", syserr)
	}
}
