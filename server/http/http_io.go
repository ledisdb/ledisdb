package http

import (
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"net/http"
)

type httpContext struct {
}

type httpReader struct {
	req *http.Request
}

type httpWriter struct {
	resp *http.ResponseWriter
}

// http context

func newHttpContext() *httpContext {
	ctx := new(httpContext)
	return ctx
}

func (ctx *httpContext) addr() string {

	return ""
}

func (ctx *httpContext) release() {

}

// http reader

func newHttpReader(req *http.Request) *httpReader {
	r := new(httpReader)
	r.req = req
	return r
}

func (r *httpReader) read() ([][]byte, error) {

	return nil, nil
}

// http writer

func newHttpWriter(resp *http.ResponseWriter) *httpWriter {
	w := new(httpWriter)
	w.resp = resp
	return w
}

func (w *httpWriter) writeError(err error) {

}

func (w *httpWriter) writeStatus(status string) {

}

func (w *httpWriter) writeInteger(n int64) {

}

func (w *httpWriter) writeBulk(b []byte) {

}

func (w *httpWriter) writeArray(lst []interface{}) {

}

func (w *httpWriter) writeSliceArray(lst [][]byte) {

}

func (w *httpWriter) writeFVPairArray(lst []ledis.FVPair) {

}

func (w *httpWriter) writeScorePairArray(lst []ledis.ScorePair, withScores bool) {

}

func (w *httpWriter) writeBulkFrom(n int64, rb io.Reader) {

}

func (w *httpWriter) flush() {

}
