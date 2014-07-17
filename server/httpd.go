package server

import (
	"net/http"
	//"github.com/siddontang/go-websocket/websocket"
)

type cmdHandler struct {
	app *App
}

func (h *cmdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("cmd handler"))
}

type wsHandler struct {
	app *App
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ws handler"))
}
