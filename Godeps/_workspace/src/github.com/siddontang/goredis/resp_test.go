package goredis

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

func TestResp(t *testing.T) {
	var buf bytes.Buffer

	reader := NewRespReader(bufio.NewReader(&buf))
	writer := NewRespWriter(bufio.NewWriter(&buf))

	if err := writer.WriteCommand("SELECT", 1); err != nil {
		t.Fatal(err)
	} else {
		if reqs, err := reader.ParseRequest(); err != nil {
			t.Fatal(err)
		} else if len(reqs) != 2 {
			t.Fatal(len(reqs))
		} else if string(reqs[0]) != "SELECT" {
			t.Fatal(string(reqs[0]))
		} else if string(reqs[1]) != "1" {
			t.Fatal(string(reqs[1]))
		}
	}

	if err := writer.FlushInteger(10); err != nil {
		t.Fatal(err)
	} else {
		if n, err := Int64(reader.Parse()); err != nil {
			t.Fatal(err)
		} else if n != 10 {
			t.Fatal(n)
		}
	}

	if err := writer.FlushString("abc"); err != nil {
		t.Fatal(err)
	} else {
		if s, err := String(reader.Parse()); err != nil {
			t.Fatal(err)
		} else if s != "abc" {
			t.Fatal(s)
		}
	}

	if err := writer.FlushBulk([]byte("abc")); err != nil {
		t.Fatal(err)
	} else {
		if s, err := String(reader.Parse()); err != nil {
			t.Fatal(err)
		} else if s != "abc" {
			t.Fatal(s)
		}
	}

	ay := []interface{}{[]byte("SET"), []byte("a"), []byte("1")}
	if err := writer.FlushArray(ay); err != nil {
		t.Fatal(err)
	} else {
		if oy, err := reader.Parse(); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(oy, ay) {
			t.Fatalf("%#v", oy)
		}
	}

	e := Error("hello world")
	if err := writer.FlushError(e); err != nil {
		t.Fatal(err)
	} else {
		if ee, err := reader.Parse(); err != nil {
			t.Fatal("must error")
		} else if !reflect.DeepEqual(e, ee) {
			t.Fatal(ee)
		}
	}
}
