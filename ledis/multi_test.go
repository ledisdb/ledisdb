package ledis

import (
	"sync"
	"testing"
)

func TestMulti(t *testing.T) {
	db := getTestDB()

	key := []byte("test_multi_1")
	v1 := []byte("v1")
	v2 := []byte("v2")

	m, err := db.Multi()
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		if err := db.Set(key, v2); err != nil {
			t.Fatal(err)
		}
		wg.Done()
	}()

	if err := m.Set(key, v1); err != nil {
		t.Fatal(err)
	}

	if v, err := m.Get(key); err != nil {
		t.Fatal(err)
	} else if string(v) != string(v1) {
		t.Fatal(string(v))
	}

	m.Close()

	wg.Wait()

	if v, err := db.Get(key); err != nil {
		t.Fatal(err)
	} else if string(v) != string(v2) {
		t.Fatal(string(v))
	}

}
