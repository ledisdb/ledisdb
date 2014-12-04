package ledis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Error represents an error returned in a command reply.
type Error string

func (err Error) Error() string { return string(err) }

type Conn struct {
	cm sync.Mutex
	wm sync.Mutex
	rm sync.Mutex

	closed bool

	client *Client

	addr string

	c  net.Conn
	br *bufio.Reader
	bw *bufio.Writer

	rSize int
	wSize int

	// Scratch space for formatting argument length.
	// '*' or '$', length, "\r\n"
	lenScratch [32]byte

	// Scratch space for formatting integers and floats.
	numScratch [40]byte

	connectTimeout time.Duration
}

func NewConn(addr string) *Conn {
	co := new(Conn)
	co.addr = addr

	co.rSize = 4096
	co.wSize = 4096

	co.closed = false

	return co
}

func NewConnSize(addr string, readSize int, writeSize int) *Conn {
	co := NewConn(addr)
	co.rSize = readSize
	co.wSize = writeSize
	return co
}

func (c *Conn) Close() {
	if c.client != nil {
		c.client.put(c)
	} else {
		c.finalize()
	}
}

func (c *Conn) SetConnectTimeout(t time.Duration) {
	c.cm.Lock()
	c.connectTimeout = t
	c.cm.Unlock()
}

func (c *Conn) SetReadDeadline(t time.Time) {
	c.cm.Lock()
	if c.c != nil {
		c.c.SetReadDeadline(t)
	}
	c.cm.Unlock()
}

func (c *Conn) SetWriteDeadline(t time.Time) {
	c.cm.Lock()
	if c.c != nil {
		c.c.SetWriteDeadline(t)
	}
	c.cm.Unlock()
}

func (c *Conn) Do(cmd string, args ...interface{}) (interface{}, error) {
	if err := c.Send(cmd, args...); err != nil {
		return nil, err
	}

	return c.Receive()
}

func (c *Conn) Send(cmd string, args ...interface{}) error {
	var err error
	for i := 0; i < 2; i++ {
		if err = c.send(cmd, args...); err != nil {
			if e, ok := err.(*net.OpError); ok && strings.Contains(e.Error(), "use of closed network connection") {
				//send to a closed connection, try again
				continue
			}
		} else {
			return nil
		}
	}
	return err
}

func (c *Conn) send(cmd string, args ...interface{}) error {
	if err := c.connect(); err != nil {
		return err
	}

	c.wm.Lock()
	defer c.wm.Unlock()

	if err := c.writeCommand(cmd, args); err != nil {
		c.finalize()
		return err
	}

	if err := c.bw.Flush(); err != nil {
		c.finalize()
		return err
	}
	return nil
}

func (c *Conn) Receive() (interface{}, error) {
	c.rm.Lock()
	defer c.rm.Unlock()

	if reply, err := c.readReply(); err != nil {
		c.finalize()
		return nil, err
	} else {
		if e, ok := reply.(Error); ok {
			return reply, e
		} else {
			return reply, nil
		}
	}
}

func (c *Conn) ReceiveBulkTo(w io.Writer) error {
	c.rm.Lock()
	defer c.rm.Unlock()

	err := c.readBulkReplyTo(w)
	if err != nil {
		if _, ok := err.(Error); !ok {
			c.finalize()
		}
	}
	return err
}

func (c *Conn) finalize() {
	c.cm.Lock()
	if !c.closed {
		if c.c != nil {
			c.c.Close()
		}
		c.closed = true
	}
	c.cm.Unlock()
}

func (c *Conn) connect() error {
	c.cm.Lock()
	defer c.cm.Unlock()

	if !c.closed && c.c != nil {
		return nil
	}

	var err error
	c.c, err = net.DialTimeout(getProto(c.addr), c.addr, c.connectTimeout)
	if err != nil {
		c.c = nil
		return err
	}

	if c.br != nil {
		c.br.Reset(c.c)
	} else {
		c.br = bufio.NewReaderSize(c.c, c.rSize)
	}

	if c.bw != nil {
		c.bw.Reset(c.c)
	} else {
		c.bw = bufio.NewWriterSize(c.c, c.wSize)
	}

	return nil
}

func (c *Conn) writeLen(prefix byte, n int) error {
	c.lenScratch[len(c.lenScratch)-1] = '\n'
	c.lenScratch[len(c.lenScratch)-2] = '\r'
	i := len(c.lenScratch) - 3
	for {
		c.lenScratch[i] = byte('0' + n%10)
		i -= 1
		n = n / 10
		if n == 0 {
			break
		}
	}
	c.lenScratch[i] = prefix
	_, err := c.bw.Write(c.lenScratch[i:])
	return err
}

func (c *Conn) writeString(s string) error {
	c.writeLen('$', len(s))
	c.bw.WriteString(s)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *Conn) writeBytes(p []byte) error {
	c.writeLen('$', len(p))
	c.bw.Write(p)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *Conn) writeInt64(n int64) error {
	return c.writeBytes(strconv.AppendInt(c.numScratch[:0], n, 10))
}

func (c *Conn) writeFloat64(n float64) error {
	return c.writeBytes(strconv.AppendFloat(c.numScratch[:0], n, 'g', -1, 64))
}

func (c *Conn) writeCommand(cmd string, args []interface{}) (err error) {
	c.writeLen('*', 1+len(args))
	err = c.writeString(cmd)
	for _, arg := range args {
		if err != nil {
			break
		}
		switch arg := arg.(type) {
		case string:
			err = c.writeString(arg)
		case []byte:
			err = c.writeBytes(arg)
		case int:
			err = c.writeInt64(int64(arg))
		case int64:
			err = c.writeInt64(arg)
		case float64:
			err = c.writeFloat64(arg)
		case bool:
			if arg {
				err = c.writeString("1")
			} else {
				err = c.writeString("0")
			}
		case nil:
			err = c.writeString("")
		default:
			var buf bytes.Buffer
			fmt.Fprint(&buf, arg)
			err = c.writeBytes(buf.Bytes())
		}
	}
	return err
}

func (c *Conn) readLine() ([]byte, error) {
	p, err := c.br.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		return nil, errors.New("ledis: long response line")
	}
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, errors.New("ledis: bad response line terminator")
	}
	return p[:i], nil
}

// parseLen parses bulk string and array lengths.
func parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, errors.New("ledis: malformed length")
	}

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return -1, errors.New("ledis: illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}

// parseInt parses an integer reply.
func parseInt(p []byte) (interface{}, error) {
	if len(p) == 0 {
		return 0, errors.New("ledis: malformed integer")
	}

	var negate bool
	if p[0] == '-' {
		negate = true
		p = p[1:]
		if len(p) == 0 {
			return 0, errors.New("ledis: malformed integer")
		}
	}

	var n int64
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return 0, errors.New("ledis: illegal bytes in length")
		}
		n += int64(b - '0')
	}

	if negate {
		n = -n
	}
	return n, nil
}

var (
	okReply   interface{} = "OK"
	pongReply interface{} = "PONG"
)

func (c *Conn) readBulkReplyTo(w io.Writer) error {
	line, err := c.readLine()
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
		if nn, err = io.CopyN(w, c.br, int64(n)); err != nil {
			return err
		} else if nn != int64(n) {
			return io.ErrShortWrite
		}

		if line, err := c.readLine(); err != nil {
			return err
		} else if len(line) != 0 {
			return errors.New("ledis: bad bulk string format")
		}
		return nil
	default:
		return fmt.Errorf("ledis: not invalid bulk string type, but %c", line[0])
	}
}

func (c *Conn) readReply() (interface{}, error) {
	line, err := c.readLine()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("ledis: short response line")
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
		return parseInt(line[1:])
	case '$':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := make([]byte, n)
		_, err = io.ReadFull(c.br, p)
		if err != nil {
			return nil, err
		}
		if line, err := c.readLine(); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, errors.New("ledis: bad bulk string format")
		}
		return p, nil
	case '*':
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		r := make([]interface{}, n)
		for i := range r {
			r[i], err = c.readReply()
			if err != nil {
				return nil, err
			}
		}
		return r, nil
	}
	return nil, errors.New("ledis: unexpected response line")
}

func (c *Client) newConn(addr string) *Conn {
	co := NewConnSize(addr, c.cfg.ReadBufferSize, c.cfg.WriteBufferSize)
	co.client = c

	return co
}
