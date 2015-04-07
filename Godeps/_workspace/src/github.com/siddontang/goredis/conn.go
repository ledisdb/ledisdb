package goredis

import (
	"bufio"
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
	c  net.Conn
	br *bufio.Reader
	bw *bufio.Writer

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
	c := new(Conn)

	var err error
	c.c, err = net.Dial(getProto(addr), addr)
	if err != nil {
		return nil, err
	}

	c.br = bufio.NewReaderSize(io.TeeReader(c.c, &c.totalReadSize), readSize)
	c.bw = bufio.NewWriterSize(io.MultiWriter(c.c, &c.totalWriteSize), writeSize)

	c.respReader = NewRespReader(c.br)
	c.respWriter = NewRespWriter(c.bw)

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

func (c *Conn) SetReadDeadline(t time.Time) {
	c.c.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) {
	c.c.SetWriteDeadline(t)
}

func (c *Conn) Do(cmd string, args ...interface{}) (interface{}, error) {
	if err := c.Send(cmd, args...); err != nil {
		return nil, err
	}

	return c.Receive()
}

func (c *Conn) Send(cmd string, args ...interface{}) error {
	if err := c.respWriter.WriteCommand(cmd, args...); err != nil {
		c.Close()
		return err
	}

	return nil
}

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

func (c *Conn) ReceiveBulkTo(w io.Writer) error {
	err := c.respReader.ParseBulkTo(w)
	if err != nil {
		if _, ok := err.(Error); !ok {
			c.Close()
		}
	}
	return err
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
