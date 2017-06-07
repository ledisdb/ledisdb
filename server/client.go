package server

import (
	"bytes"
	//	"fmt"
	"io"
	"strings"
	"time"

	"github.com/siddontang/go/sync2"
	"github.com/siddontang/ledisdb/ledis"
)

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

type syncAck struct {
	id uint64
	ch chan uint64
}

type client struct {
	app *App
	ldb *ledis.Ledis

	db *ledis.DB

	remoteAddr string
	cmd        string
	args       [][]byte

	isAuthed bool

	resp responseWriter

	syncBuf bytes.Buffer

	lastLogID sync2.AtomicUint64

	// reqErr chan error

	buf bytes.Buffer

	slaveListeningAddr string
}

func newClient(app *App) *client {
	c := new(client)

	c.app = app
	c.ldb = app.ldb
	c.isAuthed = false
	c.db, _ = app.ldb.Select(0) //use default db

	return c
}

func (c *client) close() {

}

func (c *client) authEnabled() bool {
	return len(c.app.cfg.AuthPassword) > 0 || c.app.cfg.AuthMethod != nil
}

func (c *client) perform() {
	var err error

	start := time.Now()

	c.cmd = strings.ToLower(c.cmd)

	if len(c.cmd) == 0 {
		err = ErrEmptyCommand
	} else if exeCmd, ok := regCmds[c.cmd]; !ok {
		err = ErrNotFound
	} else if c.authEnabled() && !c.isAuthed && c.cmd != "auth" {
		err = ErrNotAuthenticated
	} else {
		err = exeCmd(c)
	}

	if c.app.access != nil {
		duration := time.Since(start)

		fullCmd := c.catGenericCommand()
		cost := duration.Nanoseconds() / 1000000

		truncateLen := len(fullCmd)
		if truncateLen > 256 {
			truncateLen = 256
		}

		c.app.access.Log(c.remoteAddr, cost, fullCmd[:truncateLen], err)
	}

	if err != nil {
		c.resp.writeError(err)
	}
	c.resp.flush()
	return
}

func (c *client) catGenericCommand() []byte {
	buffer := c.buf
	buffer.Reset()

	buffer.Write([]byte(c.cmd))

	for _, arg := range c.args {
		buffer.WriteByte(' ')
		buffer.Write(arg)
	}

	return buffer.Bytes()
}

func writeValue(w responseWriter, value interface{}) {
	switch v := value.(type) {
	case []interface{}:
		w.writeArray(v)
	case [][]byte:
		w.writeSliceArray(v)
	case []byte:
		w.writeBulk(v)
	case string:
		w.writeStatus(v)
	case nil:
		w.writeBulk(nil)
	case int64:
		w.writeInteger(v)
	default:
		panic("invalid value type")
	}
}
