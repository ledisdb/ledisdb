package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/client/go/ledis"
	"reflect"
	"testing"
)

func TestCmdEval(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if v, err := ledis.Strings(c.Do("eval", "return {KEYS[1],KEYS[2],ARGV[1],ARGV[2]}", 2, "key1", "key2", "first", "second")); err != nil {
		t.Fatal(err)
	} else if len(v) != 4 {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v, []string{"key1", "key2", "first", "second"}) {
		t.Fatal(fmt.Sprintf("%v", v))
	}

	if v, err := ledis.Strings(c.Do("eval", "return {KEYS[1],KEYS[2],ARGV[1],ARGV[2]}", 2, "key1", "key2", "first", "second")); err != nil {
		t.Fatal(err)
	} else if len(v) != 4 {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v, []string{"key1", "key2", "first", "second"}) {
		t.Fatal(fmt.Sprintf("%v", v))
	}
}
