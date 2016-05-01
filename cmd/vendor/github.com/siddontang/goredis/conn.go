package goredis

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type sizeWriter int64

func (s *sizeWriter) Write(p []byte) (int, error) {
	*s += sizeWriter(len(p))
	return len(p), nil
}

type Conn struct {
	c net.Conn

	respReader *RespReader
	respWriter *RespWriter

	totalReadSize  sizeWriter
	totalWriteSize sizeWriter

	closed int32
}

func Connect(addr string) (*Conn, error) {
	return ConnectWithSize(addr, 1024, 1024)
}

func ConnectWithSize(addr string, readSize int, writeSize int) (*Conn, error) {
	conn, err := net.Dial(getProto(addr), addr)
	if err != nil {
		return nil, err
	}

	return NewConnWithSize(conn, readSize, writeSize)
}

func NewConn(conn net.Conn) (*Conn, error) {
	return NewConnWithSize(conn, 1024, 1024)
}

func NewConnWithSize(conn net.Conn, readSize int, writeSize int) (*Conn, error) {
	c := new(Conn)

	c.c = conn

	br := bufio.NewReaderSize(io.TeeReader(c.c, &c.totalReadSize), readSize)
	bw := bufio.NewWriterSize(io.MultiWriter(c.c, &c.totalWriteSize), writeSize)

	c.respReader = NewRespReader(br)
	c.respWriter = NewRespWriter(bw)

	atomic.StoreInt32(&c.closed, 0)

	return c, nil

}

func (c *Conn) Close() {
	if atomic.LoadInt32(&c.closed) == 1 {
		return
	}

	c.c.Close()

	atomic.StoreInt32(&c.closed, 1)
}

func (c *Conn) isClosed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

func (c *Conn) GetTotalReadSize() int64 {
	return int64(c.totalReadSize)
}

func (c *Conn) GetTotalWriteSize() int64 {
	return int64(c.totalWriteSize)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.c.SetWriteDeadline(t)
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

// Send RESP command and receive the reply
func (c *Conn) Do(cmd string, args ...interface{}) (interface{}, error) {
	if err := c.Send(cmd, args...); err != nil {
		return nil, err
	}

	return c.Receive()
}

// Send RESP command
func (c *Conn) Send(cmd string, args ...interface{}) error {
	if err := c.respWriter.WriteCommand(cmd, args...); err != nil {
		c.Close()
		return err
	}

	return nil
}

// Receive RESP reply
func (c *Conn) Receive() (interface{}, error) {
	if reply, err := c.respReader.Parse(); err != nil {
		c.Close()
		return nil, err
	} else {
		if e, ok := reply.(Error); ok {
			return reply, e
		} else {
			return reply, nil
		}
	}
}

// Receive RESP bulk string reply into writer w
func (c *Conn) ReceiveBulkTo(w io.Writer) error {
	err := c.respReader.ParseBulkTo(w)
	if err != nil {
		if _, ok := err.(Error); !ok {
			c.Close()
		}
	}
	return err
}

// Receive RESP command request, must array of bulk stirng
func (c *Conn) ReceiveRequest() ([][]byte, error) {
	return c.respReader.ParseRequest()
}

// Send RESP value, must be string, int64, []byte, error, nil or []interface{}
func (c *Conn) SendValue(v interface{}) error {
	switch v := v.(type) {
	case string:
		return c.respWriter.FlushString(v)
	case int64:
		return c.respWriter.FlushInteger(v)
	case []byte:
		return c.respWriter.FlushBulk(v)
	case []interface{}:
		return c.respWriter.FlushArray(v)
	case error:
		return c.respWriter.FlushError(v)
	case nil:
		return c.respWriter.FlushBulk(nil)
	default:
		return fmt.Errorf("invalid type %T for send RESP value", v)
	}
}

func (c *Client) newConn(addr string, pass string) (*Conn, error) {
	co, err := ConnectWithSize(addr, c.readBufferSize, c.writeBufferSize)
	if err != nil {
		return nil, err
	}

	if len(pass) > 0 {
		_, err = co.Do("AUTH", pass)
		if err != nil {
			co.Close()
			return nil, err
		}
	}

	return co, nil
}
