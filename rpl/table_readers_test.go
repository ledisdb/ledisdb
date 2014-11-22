package rpl

import (
	"testing"
)

func TestTableReaders(t *testing.T) {
	ts := make(tableReaders, 0, 10)

	for i := uint64(0); i < 10; i++ {
		t := new(tableReader)
		t.index = int64(i) + 1
		t.first = i*10 + 1
		t.last = i*10 + 10

		ts = append(ts, t)
	}

	if err := ts.check(); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 100; i++ {
		if r := ts.Search(uint64(i)); r == nil {
			t.Fatal("must hit", i)
		} else if r.index != int64((i-1)/10)+1 {
			t.Fatal("invalid index", r.index, i)
		}
	}

	if r := ts.Search(1000); r != nil {
		t.Fatal("must not hit")
	}
	if r := ts.Search(0); r != nil {
		t.Fatal("must not hit")
	}

}
