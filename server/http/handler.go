package http

import (
	"net/http"
	//"github.com/siddontang/go-websocket/websocket"
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/ledisdb/ledis"
	"strconv"
	"strings"
)

type CmdHandler struct {
	Ldb *ledis.Ledis
}

func (h *CmdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	idx, cmd, args := h.parseReqPath(r.URL.Path)
	cmdFunc := lookup(cmd)
	if cmdFunc == nil {
		h.cmdNotFound(cmd, w)
		return
	}
	var db *ledis.DB
	var err error
	if db, err = h.Ldb.Select(idx); err != nil {
		h.serverError(cmd, err, w)
		return
	}
	result, err := cmdFunc(db, args...)
	if err != nil {
		h.serverError(cmd, err, w)
		return
	}
	h.write(cmd, result, w)
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
func (h *CmdHandler) cmdNotFound(cmd string, w http.ResponseWriter) {
	result := [2]interface{}{
		false,
		fmt.Sprintf("ERR unknown command '%s'", cmd),
	}
	h.write(cmd, result, w)
}

func (h *CmdHandler) write(cmd string, result interface{}, w http.ResponseWriter) {
	m := map[string]interface{}{
		cmd: result,
	}

	buf, err := json.Marshal(&m)
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

func (h *CmdHandler) serverError(cmd string, err error, w http.ResponseWriter) {
	result := [2]interface{}{
		false,
		fmt.Sprintf("ERR %s", err.Error()),
	}
	h.write(cmd, result, w)
}

type WsHandler struct {
	Ldb *ledis.Ledis
}

func (h *WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ws handler"))
}
