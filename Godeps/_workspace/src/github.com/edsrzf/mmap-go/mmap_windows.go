// Copyright 2011 Evan Shaw. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mmap

import (
	"errors"
	"os"
	"sync"
	"syscall"
)

// mmap on Windows is a two-step process.
// First, we call CreateFileMapping to get a handle.
// Then, we call MapviewToFile to get an actual pointer into memory.
// Because we want to emulate a POSIX-style mmap, we don't want to expose
// the handle -- only the pointer. We also want to return only a byte slice,
// not a struct, so it's convenient to manipulate.

// We keep this map so that we can get back the original handle from the memory address.
var handleLock sync.Mutex
var handleMap = map[uintptr]syscall.Handle{}

func mmap(len int, prot, flags, hfile uintptr, off int64) ([]byte, error) {
	flProtect := uint32(syscall.PAGE_READONLY)
	dwDesiredAccess := uint32(syscall.FILE_MAP_READ)
	switch {
	case prot&COPY != 0:
		flProtect = syscall.PAGE_WRITECOPY
		dwDesiredAccess = syscall.FILE_MAP_COPY
	case prot&RDWR != 0:
		flProtect = syscall.PAGE_READWRITE
		dwDesiredAccess = syscall.FILE_MAP_WRITE
	}
	if prot&EXEC != 0 {
		flProtect <<= 4
		dwDesiredAccess |= syscall.FILE_MAP_EXECUTE
	}

	// TODO: Do we need to set some security attributes? It might help portability.
	h, errno := syscall.CreateFileMapping(syscall.Handle(hfile), nil, flProtect, 0, uint32(len), nil)
	if h == 0 {
		return nil, os.NewSyscallError("CreateFileMapping", errno)
	}

	addr, errno := syscall.MapViewOfFile(h, dwDesiredAccess, uint32(off>>32), uint32(off&0xFFFFFFFF), uintptr(len))
	if addr == 0 {
		return nil, os.NewSyscallError("MapViewOfFile", errno)
	}
	handleLock.Lock()
	handleMap[addr] = h
	handleLock.Unlock()

	m := MMap{}
	dh := m.header()
	dh.Data = addr
	dh.Len = len
	dh.Cap = dh.Len

	return m, nil
}

func flush(addr, len uintptr) error {
	errno := syscall.FlushViewOfFile(addr, len)
	return os.NewSyscallError("FlushViewOfFile", errno)
}

func lock(addr, len uintptr) error {
	errno := syscall.VirtualLock(addr, len)
	return os.NewSyscallError("VirtualLock", errno)
}

func unlock(addr, len uintptr) error {
	errno := syscall.VirtualUnlock(addr, len)
	return os.NewSyscallError("VirtualUnlock", errno)
}

func unmap(addr, len uintptr) error {
	flush(addr, len)
	err := syscall.UnmapViewOfFile(addr)
	if err != nil {
		return err
	}

	handleLock.Lock()
	defer handleLock.Unlock()
	handle, ok := handleMap[addr]
	if !ok {
		// should be impossible; we would've errored above
		return errors.New("unknown base address")
	}
	delete(handleMap, addr)

	e := syscall.CloseHandle(syscall.Handle(handle))
	return os.NewSyscallError("CloseHandle", e)
}
