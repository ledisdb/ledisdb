package server

import (
	"bytes"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"runtime"
	"sync"
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

	finish chan interface{}
}

type requestHandler struct {
	app *App

	async bool
	quit  chan struct{}
	jobs  *sync.WaitGroup

	reqs   chan *requestContext
	reqErr chan error

	buf bytes.Buffer
}

func newRequestContext(app *App) *requestContext {
	req := new(requestContext)

	req.app = app
	req.ldb = app.ldb
	req.db, _ = app.ldb.Select(0) //use default db

	req.compressBuf = make([]byte, 256)
	req.finish = make(chan interface{}, 1)

	return req
}

func newRequestHandler(app *App) *requestHandler {
	hdl := new(requestHandler)

	hdl.app = app

	hdl.async = false
	hdl.jobs = new(sync.WaitGroup)
	hdl.quit = make(chan struct{})

	hdl.reqs = make(chan *requestContext)
	hdl.reqErr = make(chan error)

	return hdl
}

func (h *requestHandler) asyncRun() {
	if !h.async {
		// todo ... not safe
		h.async = true
		go h.run()
	}
}

func (h *requestHandler) close() {
	if h.async {
		close(h.quit)
		h.jobs.Wait()
	}
}

func (h *requestHandler) run() {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			buf = buf[0:n]

			log.Fatal("request handler run panic %s:%v", buf, e)
		}
	}()

	h.jobs.Add(1)

	var req *requestContext
	for !h.async {
		select {
		case req = <-h.reqs:
			if req != nil {
				h.performance(req)
			}
		case <-h.quit:
			h.async = true
			break
		}
	}

	h.jobs.Done()
	return
}

func (h *requestHandler) postRequest(req *requestContext) {
	if h.async {
		h.reqs <- req
	} else {
		h.performance(req)
	}

	<-req.finish
}

func (h *requestHandler) performance(req *requestContext) {
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

	req.finish <- nil
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
