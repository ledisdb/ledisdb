package rpl

import (
	"container/list"
	"encoding/binary"
)

type logLRUCache struct {
	itemsList  *list.List
	itemsMap   map[uint64]*list.Element
	size       int
	capability int
	maxNum     int
}

func newLogLRUCache(capability int, maxNum int) *logLRUCache {
	if capability <= 0 {
		capability = 1024 * 1024
	}

	if maxNum <= 0 {
		maxNum = 16
	}

	return &logLRUCache{
		itemsList:  list.New(),
		itemsMap:   make(map[uint64]*list.Element),
		size:       0,
		capability: capability,
		maxNum:     maxNum,
	}
}

func (cache *logLRUCache) Set(id uint64, data []byte) {
	elem, ok := cache.itemsMap[id]
	if ok {
		//we may not enter here
		// item already exists, so move it to the front of the list and update the data
		cache.itemsList.MoveToFront(elem)
		ol := elem.Value.([]byte)
		elem.Value = data
		cache.size += (len(data) - len(ol))
	} else {
		cache.size += len(data)

		// item doesn't exist, so add it to front of list
		elem = cache.itemsList.PushFront(data)
		cache.itemsMap[id] = elem
	}

	// evict LRU entry if the cache is full
	for cache.size > cache.capability || cache.itemsList.Len() > cache.maxNum {
		removedElem := cache.itemsList.Back()
		l := removedElem.Value.([]byte)
		cache.itemsList.Remove(removedElem)
		delete(cache.itemsMap, binary.BigEndian.Uint64(l[0:8]))

		cache.size -= len(l)
		if cache.size <= 0 {
			cache.size = 0
		}
	}
}

func (cache *logLRUCache) Get(id uint64) []byte {
	elem, ok := cache.itemsMap[id]
	if !ok {
		return nil
	}

	// item exists, so move it to front of list and return it
	cache.itemsList.MoveToFront(elem)
	l := elem.Value.([]byte)
	return l
}

func (cache *logLRUCache) Delete(id uint64) {
	elem, ok := cache.itemsMap[id]
	if !ok {
		return
	}

	cache.itemsList.Remove(elem)
	delete(cache.itemsMap, id)
}

func (cache *logLRUCache) Len() int {
	return cache.itemsList.Len()
}

func (cache *logLRUCache) Reset() {
	cache.itemsList = list.New()
	cache.itemsMap = make(map[uint64]*list.Element)
	cache.size = 0
}
