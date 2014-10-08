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

	defer os.Remove("./config.toml.bak")
	if err := cfg.DumpFile("./config.toml.bak"); err != nil {
		t.Fatal(err)
	}

	if c, err := NewConfigWithFile("./config.toml.bak"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(cfg, c) {
		t.Fatal("must equal")
	}
}
