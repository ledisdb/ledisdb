package ledis

import (
	"testing"
)

func TestClient(t *testing.T) {
	cfg := new(Config)
	cfg.Addr = "127.0.0.1:6380"
	cfg.MaxIdleConns = 4

	c := NewClient(cfg)

	c.Close()
}
