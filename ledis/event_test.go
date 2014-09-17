package ledis

import (
	"reflect"
	"testing"
)

func TestEvent(t *testing.T) {
	k1 := []byte("k1")
	v1 := []byte("v1")
	k2 := []byte("k2")
	k3 := []byte("k3")
	v3 := []byte("v3")

	b := new(eventBatch)

	b.Put(k1, v1)
	b.Delete(k2)
	b.Put(k3, v3)

	buf := b.Bytes()

	ev2 := []event{
		event{k1, v1},
		event{k2, nil},
		event{k3, v3},
	}

	if ev, err := decodeEventBatch(buf); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ev, ev2) {
		t.Fatal("not equal")
	}
}
