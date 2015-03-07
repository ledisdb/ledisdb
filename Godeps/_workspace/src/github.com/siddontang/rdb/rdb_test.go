package rdb

import (
	"reflect"
	"testing"
)

func TestCodec(t *testing.T) {
	testCodec(String("abc"), t)
}

func testCodec(obj interface{}, t *testing.T) {
	b, err := Dump(obj)
	if err != nil {
		t.Fatal(err)
	}

	if o, err := DecodeDump(b); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(obj, o) {
		t.Fatal("must equal")
	}
}
