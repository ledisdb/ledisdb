package http

import (
	"net/http"
	//"github.com/siddontang/go-websocket/websocket"
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/ugorji/go/codec"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type CmdHandler struct {
	Ldb *ledis.Ledis
}

var allowedContentTypes = map[string]struct{}{
	"json":    struct{}{},
	"bson":    struct{}{},
	"msgpack": struct{}{},
}

func (h *CmdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	idx, cmd, args := h.parseReqPath(r.URL.Path)

	contentType := r.FormValue("type")
	if contentType == "" {
		contentType = "json"
	}
	contentType = strings.ToLower(contentType)
	if _, ok := allowedContentTypes[contentType]; !ok {
		h.writeError(
			cmd,
			fmt.Errorf("unsupported content type '%s', only json, bson, msgpack are supported", contentType),
			w,
			"json")
		return
	}
	cmdFunc := lookup(cmd)
	if cmdFunc == nil {
		h.cmdNotFound(cmd, w, contentType)
		return
	}
	var db *ledis.DB
	var err error
	if db, err = h.Ldb.Select(idx); err != nil {
		h.writeError(cmd, err, w, contentType)
		return
	}
	result, err := cmdFunc(db, args...)
	if err != nil {
		h.writeError(cmd, err, w, contentType)
		return
	}
	h.write(cmd, result, w, contentType)
}

func (h *CmdHandler) parseReqPath(path string) (db int, cmd string, args []string) {
	/*
	      this function extracts `db`, `cmd` and `args` from `path`
	      the proper format of `path`  is /cmd/arg1/arg2/../argN  or  /db/cmd/arg1/arg2/../argN
	   	  if `path` is the first kind, `db` will be 0
	*/
	substrings := strings.Split(strings.TrimLeft(path, "/"), "/")
	if len(substrings) == 1 {
		return 0, substrings[0], substrings[1:]
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
func (h *CmdHandler) cmdNotFound(cmd string, w http.ResponseWriter, contentType string) {
	err := fmt.Errorf("unknown command '%s'", cmd)
	h.writeError(cmd, err, w, contentType)
}

func (h *CmdHandler) write(cmd string, result interface{}, w http.ResponseWriter, contentType string) {
	m := map[string]interface{}{
		cmd: result,
	}

	switch contentType {
	case "json":
		writeJSON(&m, w)
	case "bson":
		writeBSON(&m, w)
	case "msgpack":
		writeMsgPack(&m, w)
	default:
		log.Error("invalid content type %s", contentType)
	}
}

func (h *CmdHandler) writeError(cmd string, err error, w http.ResponseWriter, contentType string) {
	result := [2]interface{}{
		false,
		fmt.Sprintf("ERR %s", err.Error()),
	}
	h.write(cmd, result, w, contentType)
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

type WsHandler struct {
	Ldb *ledis.Ledis
}

func (h *WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ws handler"))
}
