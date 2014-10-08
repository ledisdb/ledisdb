package rpl

import (
	"bytes"
	"reflect"
	"testing"
)

func TestLog(t *testing.T) {
	l1 := &Log{ID: 1, CreateTime: 100, Data: []byte("hello world")}

	var buf bytes.Buffer

	if err := l1.Encode(&buf); err != nil {
		t.Fatal(err)
	}

	l2 := &Log{}

	if err := l2.Decode(&buf); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(l1, l2) {
		t.Fatal("must equal")
	}

	if buf, err := l1.Marshal(); err != nil {
		t.Fatal(err)
	} else {
		if err = l2.Unmarshal(buf); err != nil {
			t.Fatal(err)
		}
	}

	if !reflect.DeepEqual(l1, l2) {
		t.Fatal("must equal")
	}
}
