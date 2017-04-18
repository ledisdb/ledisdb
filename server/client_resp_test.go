package server

import (
	"bufio"
	"bytes"
	"errors"
	"testing"
)

func TestRespWriter(t *testing.T) {
	for _, fixture := range []struct {
		v interface{}
		e string
	}{
		{
			v: errors.New("Some error"),
			e: "-Some error\r\n", // as described at http://redis.io/topics/protocol
		},
		{
			v: "Some status",
			e: "+Some status\r\n",
		},
		{
			v: int64(42),
			e: ":42\r\n",
		},
		{
			v: []byte("ultimate answer"),
			e: "$15\r\nultimate answer\r\n",
		},
		{
			v: []interface{}{[]byte("aaa"), []byte("bbb"), int64(42)},
			e: "*3\r\n$3\r\naaa\r\n$3\r\nbbb\r\n:42\r\n",
		},
		{
			v: [][]byte{[]byte("test"), nil, []byte("zzz")},
			e: "*3\r\n$4\r\ntest\r\n$-1\r\n$3\r\nzzz\r\n",
		},
		{
			v: nil,
			e: "$-1\r\n",
		},
		{
			v: []interface{}{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{"Foo", errors.New("Bar")}},
			e: "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n",
		},
	} {
		w := new(respWriter)
		var b bytes.Buffer
		w.buff = bufio.NewWriter(&b)
		switch v := fixture.v.(type) {
		case error:
			w.writeError(v)
		case string:
			w.writeStatus(v)
		case int64:
			w.writeInteger(v)
		case []byte:
			w.writeBulk(v)
		case []interface{}:
			w.writeArray(v)
		case [][]byte:
			w.writeSliceArray(v)
		default:
			w.writeBulk(b.Bytes())
		}
		w.flush()
		if b.String() != fixture.e {
			t.Errorf("respWriter, actual: %q, expected: %q", b.String(), fixture.e)
		}
	}

}
