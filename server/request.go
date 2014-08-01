package server

import (
	"bytes"
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"time"
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

type requestContext struct {
	app *App
	ldb *ledis.Ledis
	db  *ledis.DB

	remoteAddr string
	cmd        string
	args       [][]byte

	resp responseWriter

	syncBuf     bytes.Buffer
	compressBuf []byte
}

type requestHandler struct {
	app *App

	reqErr chan error

	buf bytes.Buffer
}

func newRequestContext(app *App) *requestContext {
	req := new(requestContext)

	req.app = app
	req.ldb = app.ldb
	req.db, _ = app.ldb.Select(0) //use default db

	req.compressBuf = make([]byte, 256)

	return req
}

func newRequestHandler(app *App) *requestHandler {
	hdl := new(requestHandler)

	hdl.app = app
	hdl.reqErr = make(chan error)

	return hdl
}

func (h *requestHandler) handle(req *requestContext) {
	var err error

	start := time.Now()

	if len(req.cmd) == 0 {
		err = ErrEmptyCommand
	} else if exeCmd, ok := regCmds[req.cmd]; !ok {
		err = ErrNotFound
	} else {
		go func() {
			h.reqErr <- exeCmd(req)
		}()

		err = <-h.reqErr
	}

	duration := time.Since(start)

	if h.app.access != nil {
		fullCmd := h.catGenericCommand(req)
		cost := duration.Nanoseconds() / 1000000

		h.app.access.Log(req.remoteAddr, cost, fullCmd[:256], err)
	}

	if err != nil {
		req.resp.writeError(err)
	}
	req.resp.flush()

	return
}

// func (h *requestHandler) catFullCommand(req *requestContext) []byte {
//
// 	// if strings.HasSuffix(cmd, "expire") {
// 	// 	catExpireCommand(c, buffer)
// 	// } else {
// 	// 	catGenericCommand(c, buffer)
// 	// }
//
// 	return h.catGenericCommand(req)
// }

func (h *requestHandler) catGenericCommand(req *requestContext) []byte {
	buffer := h.buf
	buffer.Reset()

	buffer.Write([]byte(req.cmd))

	nargs := len(req.args)
	for i, arg := range req.args {
		buffer.Write(arg)
		if i != nargs-1 {
			buffer.WriteByte(' ')
		}
	}

	return buffer.Bytes()
}
