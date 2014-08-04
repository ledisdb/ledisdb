package server

import (
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/ledis"
	"io"
	"net/http"
	"strconv"
	"strings"

	"encoding/json"
	"github.com/ugorji/go/codec"
	"gopkg.in/mgo.v2/bson"
)

var allowedContentTypes = map[string]struct{}{
	"json":    struct{}{},
	"bson":    struct{}{},
	"msgpack": struct{}{},
}
var unsopportedCommands = map[string]struct{}{
	"slaveof":  struct{}{},
	"fullsync": struct{}{},
	"sync":     struct{}{},
	"quit":     struct{}{},
}

type httpClient struct {
	app *App
	db  *ledis.DB
	ldb *ledis.Ledis

	resp responseWriter
	req  *requestContext
}

type httpWriter struct {
	contentType string
	cmd         string
	w           http.ResponseWriter
}

func newClientHTTP(app *App, w http.ResponseWriter, r *http.Request) {
	var err error
	c := new(httpClient)
	c.app = app
	c.ldb = app.ldb
	c.db, err = c.ldb.Select(0)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	c.req, err = c.makeRequest(app, r, w)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	c.req.perform()
}

func (c *httpClient) addr(r *http.Request) string {
	return r.RemoteAddr
}

func (c *httpClient) makeRequest(app *App, r *http.Request, w http.ResponseWriter) (*requestContext, error) {
	var err error

	db, cmd, argsStr, contentType := c.parseReqPath(r)

	c.db, err = app.ldb.Select(db)
	if err != nil {
		return nil, err
	}

	contentType = strings.ToLower(contentType)

	if _, ok := allowedContentTypes[contentType]; !ok {
		return nil, fmt.Errorf("unsupported content type: '%s', only json, bson, msgpack are supported", contentType)
	}

	req := newRequestContext(app)
	args := make([][]byte, len(argsStr))
	for i, arg := range argsStr {
		args[i] = []byte(arg)
	}

	req.cmd = strings.ToLower(cmd)
	if _, ok := unsopportedCommands[req.cmd]; ok {
		return nil, fmt.Errorf("unsupported command: '%s'", cmd)
	}

	req.args = args

	req.remoteAddr = c.addr(r)
	req.resp = &httpWriter{contentType, cmd, w}
	return req, nil
}

func (c *httpClient) parseReqPath(r *http.Request) (db int, cmd string, args []string, contentType string) {

	contentType = r.FormValue("type")
	if contentType == "" {
		contentType = "json"
	}

	substrings := strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")
	if len(substrings) == 1 {
		return 0, substrings[0], substrings[1:], contentType
	}
	db, err := strconv.Atoi(substrings[0])
	if err != nil {
		cmd = substrings[0]
		args = substrings[1:]
	} else {
		cmd = substrings[1]
		args = substrings[2:]
	}

	return
}

// http writer

func (w *httpWriter) genericWrite(result interface{}) {

	m := map[string]interface{}{
		w.cmd: result,
	}
	switch w.contentType {
	case "json":
		writeJSON(&m, w.w)
	case "bson":
		writeBSON(&m, w.w)
	case "msgpack":
		writeMsgPack(&m, w.w)
	default:
		log.Error("invalid content type %s", w.contentType)
	}
}

func (w *httpWriter) writeError(err error) {
	result := [2]interface{}{
		false,
		fmt.Sprintf("ERR %s", err.Error()),
	}
	w.genericWrite(result)
}

func (w *httpWriter) writeStatus(status string) {
	var success bool
	if status == OK || status == PONG {
		success = true
	}
	w.genericWrite([]interface{}{success, status})
}

func (w *httpWriter) writeInteger(n int64) {
	w.genericWrite(n)
}

func (w *httpWriter) writeBulk(b []byte) {
	if b == nil {
		w.genericWrite(nil)
	} else {
		w.genericWrite(ledis.String(b))
	}
}

func (w *httpWriter) writeArray(lst []interface{}) {
	w.genericWrite(lst)
}

func (w *httpWriter) writeSliceArray(lst [][]byte) {
	arr := make([]interface{}, len(lst))
	for i, elem := range lst {
		if elem == nil {
			arr[i] = nil
		} else {
			arr[i] = ledis.String(elem)
		}
	}
	w.genericWrite(arr)
}

func (w *httpWriter) writeFVPairArray(lst []ledis.FVPair) {
	m := make(map[string]string)
	for _, elem := range lst {
		m[ledis.String(elem.Field)] = ledis.String(elem.Value)
	}
	w.genericWrite(m)
}

func (w *httpWriter) writeScorePairArray(lst []ledis.ScorePair, withScores bool) {
	var arr []string
	if withScores {
		arr = make([]string, 2*len(lst))
		for i, data := range lst {
			arr[2*i] = ledis.String(data.Member)
			arr[2*i+1] = strconv.FormatInt(data.Score, 10)
		}
	} else {
		arr = make([]string, len(lst))
		for i, data := range lst {
			arr[i] = ledis.String(data.Member)
		}
	}
	w.genericWrite(arr)
}

func (w *httpWriter) writeBulkFrom(n int64, rb io.Reader) {
	w.writeError(fmt.Errorf("unsuport"))
}

func (w *httpWriter) flush() {

}

func writeJSON(resutl interface{}, w http.ResponseWriter) {
	buf, err := json.Marshal(resutl)
	if err != nil {
		log.Error(err.Error())
		return
	}

	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))

	_, err = w.Write(buf)
	if err != nil {
		log.Error(err.Error())
	}
}

func writeBSON(result interface{}, w http.ResponseWriter) {
	buf, err := bson.Marshal(result)
	if err != nil {
		log.Error(err.Error())
		return
	}

	w.Header().Set("Content-type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))

	_, err = w.Write(buf)
	if err != nil {
		log.Error(err.Error())
	}
}

func writeMsgPack(result interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-type", "application/octet-stream")

	var mh codec.MsgpackHandle
	enc := codec.NewEncoder(w, &mh)
	if err := enc.Encode(result); err != nil {
		log.Error(err.Error())
	}
}
