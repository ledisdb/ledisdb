package rpl

import (
	"testing"
)

func TestLogLRUCache(t *testing.T) {
	c := newLogLRUCache(180, 10)

	var i uint64
	for i = 1; i <= 10; i++ {
		l := &Log{i, 0, 0, []byte("0")}
		b, _ := l.Marshal()
		c.Set(l.ID, b)
	}

	for i = 1; i <= 10; i++ {
		if l := c.Get(i); l == nil {
			t.Fatal("must exist", i)
		}
	}

	for i = 11; i <= 20; i++ {
		l := &Log{i, 0, 0, []byte("0")}
		b, _ := l.Marshal()
		c.Set(l.ID, b)
	}

	for i = 1; i <= 10; i++ {
		if l := c.Get(i); l != nil {
			t.Fatal("must not exist", i)
		}
	}

	c.Get(11)

	l := &Log{21, 0, 0, []byte("0")}
	b, _ := l.Marshal()
	c.Set(l.ID, b)

	if l := c.Get(12); l != nil {
		t.Fatal("must nil", 12)
	}

	if l := c.Get(11); l == nil {
		t.Fatal("must not nil", 11)
	}
}
