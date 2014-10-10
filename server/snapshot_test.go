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
	cfg := new(config.Config)
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
	}

	if len(s.names) != 1 {
		t.Fatal("mut one snapshot")
	}

	s.Close()
}
