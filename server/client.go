package server

import (
	"bytes"
	"errors"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"net"
	"runtime"
	"strings"
	"time"
)

var errReadRequest = errors.New("invalid request protocol")

type client struct {
	app *App
	ldb *ledis.Ledis

	db *ledis.DB

	ctx  clientContext
	resp responseWriter
	req  requestReader

	cmd  string
	args [][]byte

	reqC chan error

	compressBuf []byte
	syncBuf     bytes.Buffer
	logBuf      bytes.Buffer
}

type clientContext interface {
	addr() string
	release()
}

type requestReader interface {
	// readLine func() ([]byte, error)
	read() ([][]byte, error)
}

type responseWriter interface {
	writeError(error)
	writeStatus(string)
	writeInteger(int64)
	writeBulk([]byte)
	writeArray([]interface{})
	writeSliceArray([][]byte)
	writeFVPairArray([]ledis.FVPair)
	writeScorePairArray([]ledis.ScorePair, bool)
	writeBulkFrom(int64, io.Reader)
	flush()
}

func newClient(app *App) *client {
	c := new(client)

	c.app = app
	c.ldb = app.ldb
	c.db, _ = app.ldb.Select(0)

	c.reqC = make(chan error, 1)

	c.compressBuf = make([]byte, 256)

	return c
}

func (c *client) run() {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			buf = buf[0:n]

			log.Fatal("client run panic %s:%v", buf, e)
		}

		c.ctx.release()
	}()

	for {
		req, err := c.req.read()
		if err != nil {
			return
		}

		c.handleRequest(req)
	}
}

func (c *client) handleRequest(req [][]byte) {
	var err error

	start := time.Now()

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

	duration := time.Since(start)

	if c.app.access != nil {
		c.logBuf.Reset()
		for i, r := range req {
			left := 256 - c.logBuf.Len()
			if left <= 0 {
				break
			} else if len(r) <= left {
				c.logBuf.Write(r)
				if i != len(req)-1 {
					c.logBuf.WriteByte(' ')
				}
			} else {
				c.logBuf.Write(r[0:left])
			}
		}

		c.app.access.Log(c.ctx.addr(), duration.Nanoseconds()/1000000, c.logBuf.Bytes(), err)
	}

	if err != nil {
		c.resp.writeError(err)
	}

	c.resp.flush()
}

func newTcpClient(conn net.Conn, app *App) {
	c := newClient(app)

	c.ctx = newTcpContext(conn)
	c.req = newTcpReader(conn)
	c.resp = newTcpWriter(conn)

	go c.run()
}

// func newHttpClient(w http.ResponseWriter, r *http.Request, app *App) {
// 	c := newClient(app)
// 	go c.run()
// }
