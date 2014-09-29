package ledis

import (
	"reflect"
	"testing"
)

type testEvent struct {
	Key   []byte
	Value []byte
}

type testEventWriter struct {
	evs []testEvent
}

func (w *testEventWriter) Put(key []byte, value []byte) {
	e := testEvent{key, value}
	w.evs = append(w.evs, e)
}

func (w *testEventWriter) Delete(key []byte) {
	e := testEvent{key, nil}
	w.evs = append(w.evs, e)
}

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

	w := &testEventWriter{}

	ev2 := &testEventWriter{
		evs: []testEvent{
			testEvent{k1, v1},
			testEvent{k2, nil},
			testEvent{k3, v3}},
	}

	if err := decodeEventBatch(w, buf); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(w, ev2) {
		t.Fatal("not equal")
	}
}
