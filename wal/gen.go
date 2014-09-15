package wal

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"sync"
)

type FileIDGenerator struct {
	LogIDGenerator

	m sync.Mutex
	f *os.File

	id uint64
}

func NewFileIDGenerator(base string) (*FileIDGenerator, error) {
	if err := os.MkdirAll(base, 0755); err != nil {
		return nil, err
	}

	g := new(FileIDGenerator)

	name := path.Join(base, "log.id")

	var err error
	if g.f, err = os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return nil, err
	}

	s, _ := g.f.Stat()
	if s.Size() == 0 {
		g.id = 0
	} else if s.Size() == 8 {
		if err = binary.Read(g.f, binary.BigEndian, &g.id); err != nil {
			g.f.Close()
			return nil, err
		} else if g.id == InvalidLogID {
			g.f.Close()
			return nil, fmt.Errorf("read invalid log id in %s", name)
		}
	} else {
		g.f.Close()
		return nil, fmt.Errorf("log id file %s is invalid", name)
	}

	return g, nil
}

func (g *FileIDGenerator) Reset(id uint64) error {
	g.m.Lock()
	defer g.m.Unlock()

	if g.f == nil {
		return fmt.Errorf("generator closed")
	}

	if g.id < id {
		g.id = id
	}

	return nil
}

func (g *FileIDGenerator) GenerateID() (uint64, error) {
	g.m.Lock()
	defer g.m.Unlock()

	if g.f == nil {
		return 0, fmt.Errorf("generator closed")
	}

	if _, err := g.f.Seek(0, os.SEEK_SET); err != nil {
		return 0, nil
	}

	id := g.id + 1

	if err := binary.Write(g.f, binary.BigEndian, id); err != nil {
		return 0, nil
	}

	g.id = id

	return id, nil
}

func (g *FileIDGenerator) Close() error {
	g.m.Lock()
	defer g.m.Unlock()

	if g.f != nil {
		err := g.f.Close()
		g.f = nil
		return err
	}
	return nil
}

type MemIDGenerator struct {
	m sync.Mutex

	LogIDGenerator

	id uint64
}

func NewMemIDGenerator(baseID uint64) *MemIDGenerator {
	g := &MemIDGenerator{id: baseID}
	return g
}

func (g *MemIDGenerator) Reset(id uint64) error {
	g.m.Lock()
	defer g.m.Unlock()

	if g.id < id {
		g.id = id
	}
	return nil
}

func (g *MemIDGenerator) GenerateID() (uint64, error) {
	g.m.Lock()
	defer g.m.Unlock()

	g.id++
	id := g.id
	return id, nil
}

func (g *MemIDGenerator) Close() error {
	return nil
}
