package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	_, err := NewConfigWithFile("./config.toml")
	if err != nil {
		t.Fatal(err)
	}

}
