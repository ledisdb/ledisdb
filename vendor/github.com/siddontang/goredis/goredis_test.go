package goredis

import (
	"github.com/alicebob/miniredis"
	"testing"
)

func Test(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	s.RequireAuth("123456")

	addr := s.Addr()

	c := NewClient(addr, "123456")
	defer c.Close()

	conn, err := c.Get()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if pong, err := String(conn.Do("PING")); err != nil {
		t.Fatal(err)
	} else if pong != "PONG" {
		t.Fatal(pong)
	}

	if pong, err := String(conn.Do("PING")); err != nil {
		t.Fatal(err)
	} else if pong != "PONG" {
		t.Fatal(pong)
	}
}
