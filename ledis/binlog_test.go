package ledis

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestBinLog(t *testing.T) {
	cfg := new(BinLogConfig)

	cfg.MaxFileNum = 1
	cfg.MaxFileSize = 1024
	cfg.Path = "/tmp/ledis_binlog"
	cfg.Name = "ledis"

	os.RemoveAll(cfg.Path)

	b, err := NewBinLogWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := b.Log(make([]byte, 1024)); err != nil {
		t.Fatal(err)
	}

	if err := b.Log(make([]byte, 1024)); err != nil {
		t.Fatal(err)
	}

	if fs, err := ioutil.ReadDir(cfg.Path); err != nil {
		t.Fatal(err)
	} else if len(fs) != 2 {
		t.Fatal(len(fs))
	}
}
