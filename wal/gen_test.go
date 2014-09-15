package wal

import (
	"io/ioutil"
	"os"
	"testing"
)

func testGenerator(t *testing.T, g LogIDGenerator, base uint64) {
	for i := base; i < base+100; i++ {
		id, err := g.GenerateID()
		if err != nil {
			t.Fatal(err)
		} else if id != i {
			t.Fatal(id, i)
		}
	}
}

func TestGenerator(t *testing.T) {
	base, err := ioutil.TempDir("", "wal")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(base)

	var g *FileIDGenerator
	if g, err = NewFileIDGenerator(base); err != nil {
		t.Fatal(err)
	} else {
		testGenerator(t, g, 1)
		if err = g.Close(); err != nil {
			t.Fatal(err)
		}
	}

	if g, err = NewFileIDGenerator(base); err != nil {
		t.Fatal(err)
	} else {
		testGenerator(t, g, 101)
		if err = g.Close(); err != nil {
			t.Fatal(err)
		}
	}

	m := NewMemIDGenerator(100)
	testGenerator(t, m, 101)
}
