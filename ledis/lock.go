package ledis

import (
	"hash/crc32"
	"sync"
)

type keyMutex struct {
	mutexs []*sync.Mutex
}

func newKeyMutex(size int) *keyMutex {
	m := new(keyMutex)

	m.mutexs = make([]*sync.Mutex, size)

	for i := range m.mutexs {
		m.mutexs[i] = &sync.Mutex{}
	}

	return m
}

func (k *keyMutex) Get(key []byte) *sync.Mutex {
	h := int(crc32.ChecksumIEEE(key))
	return k.mutexs[h%len(k.mutexs)]
}
