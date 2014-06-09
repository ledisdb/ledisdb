package server

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"net"
	"runtime"
	"strconv"
	"strings"
)

var errReadRequest = errors.New("invalid request protocol")

type client struct {
	app *App
	ldb *ledis.Ledis

	db *ledis.DB
	c  net.Conn

	rb *bufio.Reader
	wb *bufio.Writer

	cmd  string
	args [][]byte

	reqC chan error

	syncBuf bytes.Buffer
}

func newClient(c net.Conn, app *App) {
	co := new(client)

	co.app = app
	co.ldb = app.ldb
	//use default db
	co.db, _ = app.ldb.Select(0)
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
	return readLine(c.rb)
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
	if nparams, err = strconv.Atoi(ledis.String(l[1:])); err != nil {
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
			if n, err = strconv.Atoi(ledis.String(l[1:])); err != nil {
				return nil, err
			} else if n == -1 {
				req = append(req, nil)
			} else {
				buf := make([]byte, n)
				if _, err = io.ReadFull(c.rb, buf); err != nil {
					return nil, err
				}

				if l, err = c.readLine(); err != nil {
					return nil, err
				} else if len(l) != 0 {
					return nil, errors.New("bad bulk string format")
				}

				req = append(req, buf)

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
		c.cmd = strings.ToLower(ledis.String(req[0]))
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
	c.wb.Write(ledis.Slice("-ERR"))
	if err != nil {
		c.wb.WriteByte(' ')
		c.wb.Write(ledis.Slice(err.Error()))
	}
	c.wb.Write(Delims)
}

func (c *client) writeStatus(status string) {
	c.wb.WriteByte('+')
	c.wb.Write(ledis.Slice(status))
	c.wb.Write(Delims)
}

func (c *client) writeInteger(n int64) {
	c.wb.WriteByte(':')
	c.wb.Write(ledis.StrPutInt64(n))
	c.wb.Write(Delims)
}

func (c *client) writeBulk(b []byte) {
	c.wb.WriteByte('$')
	if b == nil {
		c.wb.Write(NullBulk)
	} else {
		c.wb.Write(ledis.Slice(strconv.Itoa(len(b))))
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
		c.wb.Write(ledis.Slice(strconv.Itoa(len(ay))))
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

func (c *client) writeBulkFrom(n int64, rb io.Reader) {
	c.wb.WriteByte('$')
	c.wb.Write(ledis.Slice(strconv.FormatInt(n, 10)))
	c.wb.Write(Delims)

	io.Copy(c.wb, rb)
	c.wb.Write(Delims)
}
