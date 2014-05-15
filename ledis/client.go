package ledis

import (
	"bufio"
	"errors"
	"github.com/siddontang/golib/log"
	"io"
	"net"
	"runtime"
	"strconv"
	"strings"
)

var errReadRequest = errors.New("invalid request protocol")

type client struct {
	db *DB
	c  net.Conn

	rb *bufio.Reader
	wb *bufio.Writer

	cmd  string
	args [][]byte

	reqC chan error
}

func newClient(c net.Conn, db *DB) {
	co := new(client)
	co.db = db
	co.c = c

	co.rb = bufio.NewReaderSize(c, 256)
	co.wb = bufio.NewWriterSize(c, 256)

	co.reqC = make(chan error, 1)

	go co.run()
}

func (c *client) run() {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			buf = buf[0:n]

			log.Fatal("client run panic %s:%v", buf, e)
		}

		c.c.Close()
	}()

	for {
		req, err := c.readRequest()
		if err != nil {
			return
		}

		c.handleRequest(req)
	}
}

func (c *client) readLine() ([]byte, error) {
	var line []byte
	for {
		l, more, err := c.rb.ReadLine()
		if err != nil {
			return nil, err
		}

		if line == nil && !more {
			return l, nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}

//A client sends to the Redis server a RESP Array consisting of just Bulk Strings.
func (c *client) readRequest() ([][]byte, error) {
	l, err := c.readLine()
	if err != nil {
		return nil, err
	} else if len(l) == 0 || l[0] != '*' {
		return nil, errReadRequest
	}

	var nparams int
	if nparams, err = strconv.Atoi(String(l[1:])); err != nil {
		return nil, err
	} else if nparams <= 0 {
		return nil, errReadRequest
	}

	req := make([][]byte, 0, nparams)
	var n int
	for i := 0; i < nparams; i++ {
		if l, err = c.readLine(); err != nil {
			return nil, err
		}

		if len(l) == 0 {
			return nil, errReadRequest
		} else if l[0] == '$' {
			//handle resp string
			if n, err = strconv.Atoi(String(l[1:])); err != nil {
				return nil, err
			} else if n == -1 {
				req = append(req, nil)
			} else {
				buf := make([]byte, n+2)
				if _, err = io.ReadFull(c.rb, buf); err != nil {
					return nil, err
				} else if buf[len(buf)-2] != '\r' || buf[len(buf)-1] != '\n' {
					return nil, errReadRequest

				} else {
					req = append(req, buf[0:len(buf)-2])
				}
			}

		} else {
			return nil, errReadRequest
		}
	}

	return req, nil
}

func (c *client) handleRequest(req [][]byte) {
	var err error

	if len(req) == 0 {
		err = ErrEmptyCommand
	} else {
		c.cmd = strings.ToLower(String(req[0]))
		c.args = req[1:]

		f, ok := regCmds[c.cmd]
		if !ok {
			err = ErrNotFound
		} else {
			go func() {
				c.reqC <- f(c)
			}()
			err = <-c.reqC
		}
	}

	if err != nil {
		c.writeError(err)
	}

	c.wb.Flush()
}

func (c *client) writeError(err error) {
	c.wb.Write(Slice("-ERR"))
	if err != nil {
		c.wb.WriteByte(' ')
		c.wb.Write(Slice(err.Error()))
	}
	c.wb.Write(Delims)
}

func (c *client) writeStatus(status string) {
	c.wb.WriteByte('+')
	c.wb.Write(Slice(status))
	c.wb.Write(Delims)
}

func (c *client) writeInteger(n int64) {
	c.wb.WriteByte(':')
	c.wb.Write(Slice(strconv.FormatInt(n, 10)))
	c.wb.Write(Delims)
}

func (c *client) writeBulk(b []byte) {
	c.wb.WriteByte('$')
	if b == nil {
		c.wb.Write(NullBulk)
	} else {
		c.wb.Write(Slice(strconv.Itoa(len(b))))
		c.wb.Write(Delims)
		c.wb.Write(b)
	}

	c.wb.Write(Delims)
}

func (c *client) writeArray(ay []interface{}) {
	c.wb.WriteByte('*')
	if ay == nil {
		c.wb.Write(NullArray)
		c.wb.Write(Delims)
	} else {
		c.wb.Write(Slice(strconv.Itoa(len(ay))))
		c.wb.Write(Delims)

		for i := 0; i < len(ay); i++ {
			switch v := ay[i].(type) {
			case []interface{}:
				c.writeArray(v)
			case []byte:
				c.writeBulk(v)
			case nil:
				c.writeBulk(nil)
			case int64:
				c.writeInteger(v)
			default:
				panic("invalid array type")
			}
		}
	}
}
