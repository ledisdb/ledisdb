package rpl

import (
	"github.com/siddontang/ledisdb/config"
	"io/ioutil"
	"os"
	"testing"
)

func TestReplication(t *testing.T) {
	dir, err := ioutil.TempDir("", "rpl")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer os.RemoveAll(dir)

	c := config.NewConfigDefault()
	c.Replication.Path = dir

	r, err := NewReplication(c)
	if err != nil {
		t.Fatal(err)
	}

	if l1, err := r.Log([]byte("hello world")); err != nil {
		t.Fatal(err)
	} else if l1.ID != 1 {
		t.Fatal(l1.ID)
	}

	if b, _ := r.CommitIDBehind(); !b {
		t.Fatal("must backward")
	}

	if err := r.UpdateCommitID(1); err != nil {
		t.Fatal(err)
	}

	if b, _ := r.CommitIDBehind(); b {
		t.Fatal("must not backward")
	}

	r.Close()
}
