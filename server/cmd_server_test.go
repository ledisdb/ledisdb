package server

import (
	"testing"

	"github.com/siddontang/goredis"
)

func TestAuth(t *testing.T) {
	c1 := getTestConn()
	defer c1.Close()

	// Should error, no params
	_, err := c1.Do("AUTH")
	if err == nil {
		t.Fatal(err)
	}

	// Should error, invalid pass
	_, err = c1.Do("AUTH", "password")
	if err.Error() != "authentication failure" {
		t.Fatal("Expected authentication error:", err)
	}

	c2 := getTestConnAuth("password")
	defer c2.Close()

	// Should fail doing a command as we've not authed
	_, err = c2.Do("GET", "tmp_select_key")
	if err.Error() != "not authenticated" {
		t.Fatal("Expected authentication error:", err)
	}

	// Login
	_, err = c2.Do("AUTH", "password")
	if err != nil {
		t.Fatal(err)
	}

	// Should be ok doing a command
	_, err = c2.Do("GET", "tmp_select_key")
	if err != nil {
		t.Fatal(err)
	}

	// Log out by sending wrong pass
	_, err = c2.Do("AUTH", "wrong password")
	if err.Error() != "authentication failure" {
		t.Fatal("Expected authentication error:", err)
	}

	// Should fail doing a command as we're logged out
	_, err = c2.Do("GET", "tmp_select_key")
	if err.Error() != "not authenticated" {
		t.Fatal("Expected authentication error:", err)
	}
}

func TestXSelect(t *testing.T) {
	c1 := getTestConn()
	defer c1.Close()

	c2 := getTestConn()
	defer c2.Close()

	_, err := c1.Do("XSELECT", "1", "THEN", "SET", "tmp_select_key", "1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = goredis.Int(c2.Do("GET", "tmp_select_key"))
	if err != goredis.ErrNil {
		t.Fatal(err)
	}

	n, _ := goredis.Int(c2.Do("XSELECT", "1", "THEN", "GET", "tmp_select_key"))
	if n != 1 {
		t.Fatal(n)
	}

	n, _ = goredis.Int(c2.Do("GET", "tmp_select_key"))
	if n != 1 {
		t.Fatal(n)
	}

	c1.Do("SELECT", 0)
	c2.Do("SELECT", 0)

}
