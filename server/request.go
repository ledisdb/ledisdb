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

	reqErr chan error

	buf bytes.Buffer
}

func newRequestContext(app *App) *requestContext {
	req := new(requestContext)

	req.app = app
	req.ldb = app.ldb
	req.db, _ = app.ldb.Select(0) //use default db

	req.compressBuf = make([]byte, 256)
	req.reqErr = make(chan error)

	return req
}

func (req *requestContext) perform() {
	var err error

	start := time.Now()

	if len(req.cmd) == 0 {
		err = ErrEmptyCommand
	} else if exeCmd, ok := regCmds[req.cmd]; !ok {
		err = ErrNotFound
	} else {
		go func() {
			req.reqErr <- exeCmd(req)
		}()

		err = <-req.reqErr
	}

	duration := time.Since(start)

	if req.app.access != nil {
		fullCmd := req.catGenericCommand()
		cost := duration.Nanoseconds() / 1000000

		truncateLen := len(fullCmd)
		if truncateLen > 256 {
			truncateLen = 256
		}

		req.app.access.Log(req.remoteAddr, cost, fullCmd[:truncateLen], err)
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

func (req *requestContext) catGenericCommand() []byte {
	buffer := req.buf
	buffer.Reset()

	buffer.Write([]byte(req.cmd))

	for _, arg := range req.args {
		buffer.WriteByte(' ')
		buffer.Write(arg)
	}

	return buffer.Bytes()
}
