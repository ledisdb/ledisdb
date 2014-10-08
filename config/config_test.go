package config

import (
	"os"
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	cfg, err := NewConfigWithFile("./config.toml")
	if err != nil {
		t.Fatal(err)
	}

	bakFile := "./config.toml.bak"

	defer os.Remove(bakFile)
	if err := cfg.DumpFile(bakFile); err != nil {
		t.Fatal(err)
	}

	if c, err := NewConfigWithFile(bakFile); err != nil {
		t.Fatal(err)
	} else {
		c.FileName = cfg.FileName
		if !reflect.DeepEqual(cfg, c) {
			t.Fatal("must equal")
		}

		c.FileName = bakFile
		c.SlaveOf = "127.0.0.1:6381"
		if err := c.Rewrite(); err != nil {
			t.Fatal(err)
		}

		if c1, err := NewConfigWithFile(bakFile); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(c, c1) {
			t.Fatalf("must equal %v != %v", c, c1)
		}
	}
}
