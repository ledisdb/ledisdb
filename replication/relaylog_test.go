package replication

import (
	"os"
	"testing"
)

func TestRelayLog(t *testing.T) {
	cfg := new(RelayLogConfig)

	cfg.MaxFileSize = 1024
	cfg.SpaceLimit = 1024
	cfg.Path = "/tmp/ledis_relaylog"
	cfg.Name = "ledis"

	os.RemoveAll(cfg.Path)

	b, err := NewRelayLogWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := b.Log(make([]byte, 1024)); err != nil {
		t.Fatal(err)
	}

	if err := b.Log(make([]byte, 1)); err == nil {
		t.Fatal("must not nil")
	} else if err != ErrOverSpaceLimit {
		t.Fatal(err)
	}

	if err := b.Purge(1); err != nil {
		t.Fatal(err)
	}

	if err := b.Log(make([]byte, 1)); err != nil {
		t.Fatal(err)
	}

}
