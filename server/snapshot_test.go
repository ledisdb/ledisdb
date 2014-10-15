package server

import (
	"github.com/siddontang/ledisdb/config"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

type testSnapshotDumper struct {
}

func (d *testSnapshotDumper) Dump(w io.Writer) error {
	w.Write([]byte("hello world"))
	return nil
}

func TestSnapshot(t *testing.T) {
	cfg := config.NewConfigDefault()
	cfg.Snapshot.MaxNum = 2
	cfg.Snapshot.Path = path.Join(os.TempDir(), "snapshot")

	defer os.RemoveAll(cfg.Snapshot.Path)

	d := new(testSnapshotDumper)

	s, err := newSnapshotStore(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if f, _, err := s.Create(d); err != nil {
		t.Fatal(err)
	} else {
		defer f.Close()
		if b, _ := ioutil.ReadAll(f); string(b) != "hello world" {
			t.Fatal("invalid read snapshot")
		}

		if len(s.names) != 1 {
			t.Fatal("must 1 snapshot")
		}
	}

	if f, _, err := s.Create(d); err != nil {
		t.Fatal(err)
	} else {
		defer f.Close()
		if b, _ := ioutil.ReadAll(f); string(b) != "hello world" {
			t.Fatal("invalid read snapshot")
		}
		if len(s.names) != 2 {
			t.Fatal("must 2 snapshot")
		}
	}

	if f, _, err := s.Create(d); err != nil {
		t.Fatal(err)
	} else {
		defer f.Close()
		if b, _ := ioutil.ReadAll(f); string(b) != "hello world" {
			t.Fatal("invalid read snapshot")
		}

		if len(s.names) != 2 {
			t.Fatal("must 2 snapshot")
		}
	}

	fs, _ := ioutil.ReadDir(cfg.Snapshot.Path)
	if len(fs) != 2 {
		t.Fatal("must 2 snapshot")
	}

	s.Close()
}
