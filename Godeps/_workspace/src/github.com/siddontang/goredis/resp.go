package goredis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Error string

func (err Error) Error() string { return string(err) }

var (
	okReply   interface{} = "OK"
	pongReply interface{} = "PONG"
)

type RespReader struct {
	br *bufio.Reader
}

func NewRespReader(br *bufio.Reader) *RespReader {
	r := &RespReader{br}
	return r
}

// Parse RESP
func (resp *RespReader) Parse() (interface{}, error) {
	line, err := readLine(resp.br)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("short resp line")
	}
	switch line[0] {
	case '+':
		switch {
		case len(line) == 3 && line[1] == 'O' && line[2] == 'K':
			// Avoid allocation for frequent "+OK" response.
			return okReply, nil
		case len(line) == 5 && line[1] == 'P' && line[2] == 'O' && line[3] == 'N' && line[4] == 'G':
			// Avoid allocation in PING command benchmarks :)
			return pongReply, nil
		default:
			return string(line[1:]), nil
		}
	case '-':
		return Error(string(line[1:])), nil
	case ':':
		n, err := parseInt(line[1:])
		return n, err
	case '$':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := make([]byte, n)
		_, err = io.ReadFull(resp.br, p)
		if err != nil {
			return nil, err
		}
		if line, err := readLine(resp.br); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, errors.New("bad bulk string format")
		}
		return p, nil
	case '*':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		r := make([]interface{}, n)
		for i := range r {
			r[i], err = resp.Parse()
			if err != nil {
				return nil, err
			}
		}
		return r, nil
	}
	return nil, errors.New("unexpected response line")
}

// Parse client -> server command request, must be array of bulk strings
func (resp *RespReader) ParseRequest() ([][]byte, error) {
	line, err := readLine(resp.br)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("short resp line")
	}
	switch line[0] {
	case '*':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		r := make([][]byte, n)
		for i := range r {
			r[i], err = parseBulk(resp.br)
			if err != nil {
				return nil, err
			}
		}
		return r, nil
	default:
		return nil, fmt.Errorf("not invalid array of bulk string type, but %c", line[0])
	}
}

// Parse bulk string and write it with writer w
func (resp *RespReader) ParseBulkTo(w io.Writer) error {
	line, err := readLine(resp.br)
	if err != nil {
		return err
	}
	if len(line) == 0 {
		return errors.New("ledis: short response line")
	}

	switch line[0] {
	case '-':
		return Error(string(line[1:]))
	case '$':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return err
		}

		var nn int64
		if nn, err = io.CopyN(w, resp.br, int64(n)); err != nil {
			return err
		} else if nn != int64(n) {
			return io.ErrShortWrite
		}

		if line, err := readLine(resp.br); err != nil {
			return err
		} else if len(line) != 0 {
			return errors.New("bad bulk string format")
		}
		return nil
	default:
		return fmt.Errorf("not invalid bulk string type, but %c", line[0])
	}
}

func readLine(br *bufio.Reader) ([]byte, error) {
	p, err := br.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		return nil, errors.New("long resp line")
	}
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, errors.New("bad resp line terminator")
	}
	return p[:i], nil
}

// parseLen parses bulk string and array lengths.
func parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, errors.New("malformed length")
	}

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return -1, errors.New("illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}

// parseInt parses an integer reply.
func parseInt(p []byte) (int64, error) {
	if len(p) == 0 {
		return 0, errors.New("malformed integer")
	}

	var negate bool
	if p[0] == '-' {
		negate = true
		p = p[1:]
		if len(p) == 0 {
			return 0, errors.New("malformed integer")
		}
	}

	var n int64
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return 0, errors.New("illegal bytes in length")
		}
		n += int64(b - '0')
	}

	if negate {
		n = -n
	}
	return n, nil
}

func parseBulk(br *bufio.Reader) ([]byte, error) {
	line, err := readLine(br)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("short resp line")
	}
	switch line[0] {
	case '$':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := make([]byte, n)
		_, err = io.ReadFull(br, p)
		if err != nil {
			return nil, err
		}
		if line, err := readLine(br); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, errors.New("bad bulk string format")
		}
		return p, nil
	default:
		return nil, fmt.Errorf("not invalid bulk string type, but %c", line[0])
	}
}

var (
	intBuffer [][]byte
	respTerm  = []byte("\r\n")
	nullBulk  = []byte("-1")
	nullArray = []byte("-1")
)

func init() {
	cnt := 10000
	intBuffer = make([][]byte, cnt)
	for i := 0; i < cnt; i++ {
		intBuffer[i] = []byte(strconv.Itoa(i))
	}
}

type RespWriter struct {
	bw *bufio.Writer
	// Scratch space for formatting integers and floats.
	numScratch [40]byte
}

func NewRespWriter(bw *bufio.Writer) *RespWriter {
	r := &RespWriter{bw: bw}
	return r
}

func (resp *RespWriter) Flush() error {
	return resp.bw.Flush()
}

func (resp *RespWriter) writeTerm() error {
	_, err := resp.bw.Write(respTerm)
	return err
}

func (resp *RespWriter) writeInteger(n int64) error {
	var err error
	if n < int64(len(intBuffer)) {
		_, err = resp.bw.Write(intBuffer[n])
	} else {
		_, err = resp.bw.Write(strconv.AppendInt(nil, n, 10))
	}
	return err
}

func (resp *RespWriter) WriteInteger(n int64) error {
	resp.bw.WriteByte(':')

	resp.writeInteger(n)

	return resp.writeTerm()
}

func (resp *RespWriter) FlushInteger(n int64) error {
	resp.WriteInteger(n)
	return resp.Flush()
}

func (resp *RespWriter) WriteString(s string) error {
	resp.bw.WriteByte('+')
	resp.bw.WriteString(s)
	return resp.writeTerm()
}

func (resp *RespWriter) FlushString(s string) error {
	resp.WriteString(s)
	return resp.Flush()
}

func (resp *RespWriter) WriteError(e error) error {
	resp.bw.WriteByte('-')

	if e != nil {
		resp.bw.WriteString(e.Error())
	} else {
		resp.bw.WriteString("error is nil, invalid")
	}

	return resp.writeTerm()
}

func (resp *RespWriter) FlushError(e error) error {
	resp.WriteError(e)
	return resp.Flush()
}

func (resp *RespWriter) WriteBulk(b []byte) error {
	resp.bw.WriteByte('$')
	if b == nil {
		resp.bw.Write(nullBulk)
	} else {
		resp.writeInteger(int64(len(b)))
		resp.writeTerm()
		resp.bw.Write(b)
	}
	return resp.writeTerm()
}

func (resp *RespWriter) FlushBulk(b []byte) error {
	resp.WriteBulk(b)
	return resp.Flush()
}

func (resp *RespWriter) WriteArray(ay []interface{}) error {
	resp.bw.WriteByte('*')
	if ay == nil {
		resp.bw.Write(nullArray)
		return resp.writeTerm()
	} else {
		resp.writeInteger(int64(len(ay)))
		resp.writeTerm()

		var err error
		for i := 0; i < len(ay); i++ {
			if err != nil {
				return err
			}

			switch v := ay[i].(type) {
			case []interface{}:
				err = resp.WriteArray(v)
			case []byte:
				err = resp.WriteBulk(v)
			case nil:
				err = resp.WriteBulk(nil)
			case int64:
				err = resp.WriteInteger(v)
			case string:
				err = resp.WriteString(v)
			case error:
				err = resp.WriteError(v)
			default:
				err = fmt.Errorf("invalid array type %T %v", ay[i], v)
			}
		}
		return err
	}
}

func (resp *RespWriter) FlushArray(ay []interface{}) error {
	resp.WriteArray(ay)
	return resp.Flush()
}

func (resp *RespWriter) writeBulkString(s string) error {
	resp.bw.WriteByte('$')
	resp.writeInteger(int64(len(s)))
	resp.writeTerm()
	resp.bw.WriteString(s)
	return resp.writeTerm()
}

func (resp *RespWriter) writeBulkInt64(n int64) error {
	return resp.WriteBulk(strconv.AppendInt(resp.numScratch[:0], n, 10))
}

func (resp *RespWriter) writeBulkFloat64(n float64) error {
	return resp.WriteBulk(strconv.AppendFloat(resp.numScratch[:0], n, 'g', -1, 64))
}

// RESP command is array of bulk string
func (resp *RespWriter) WriteCommand(cmd string, args ...interface{}) error {
	resp.bw.WriteByte('*')

	resp.writeInteger(int64(1 + len(args)))
	resp.writeTerm()

	err := resp.writeBulkString(cmd)

	for _, arg := range args {
		if err != nil {
			break
		}
		switch arg := arg.(type) {
		case string:
			err = resp.writeBulkString(arg)
		case []byte:
			err = resp.WriteBulk(arg)
		case int:
			err = resp.writeBulkInt64(int64(arg))
		case int64:
			err = resp.writeBulkInt64(arg)
		case float64:
			err = resp.writeBulkFloat64(arg)
		case bool:
			if arg {
				err = resp.writeBulkString("1")
			} else {
				err = resp.writeBulkString("0")
			}
		case nil:
			err = resp.writeBulkString("")
		default:
			var buf bytes.Buffer
			fmt.Fprint(&buf, arg)
			err = resp.WriteBulk(buf.Bytes())
		}
	}

	if err != nil {
		return err
	}

	return resp.Flush()
}
