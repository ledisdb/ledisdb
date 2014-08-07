package server

import (
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"net"
	"net/http"
	"path"
	"strings"
)

type App struct {
	cfg *config.Config

	listener     net.Listener
	httpListener net.Listener

	ldb *ledis.Ledis

	closed bool

	quit chan struct{}

	access *accessLog

	//for slave replication
	m *master
}

func netType(s string) string {
	if strings.Contains(s, "/") {
		return "unix"
	} else {
		return "tcp"
	}
}

func NewApp(cfg *config.Config) (*App, error) {
	if len(cfg.DataDir) == 0 {
		println("use default datadir %s", config.DefaultDataDir)
		cfg.DataDir = config.DefaultDataDir
	}

	app := new(App)

	app.quit = make(chan struct{})

	app.closed = false

	app.cfg = cfg

	var err error

	if app.listener, err = net.Listen(netType(cfg.Addr), cfg.Addr); err != nil {
		return nil, err
	}

	if len(cfg.HttpAddr) > 0 {
		if app.httpListener, err = net.Listen(netType(cfg.HttpAddr), cfg.HttpAddr); err != nil {
			return nil, err
		}
	}

	if len(cfg.AccessLog) > 0 {
		if path.Dir(cfg.AccessLog) == "." {
			app.access, err = newAcessLog(path.Join(cfg.DataDir, cfg.AccessLog))
		} else {
			app.access, err = newAcessLog(cfg.AccessLog)
		}

		if err != nil {
			return nil, err
		}
	}

	if app.ldb, err = ledis.Open(cfg); err != nil {
		return nil, err
	}

	app.m = newMaster(app)

	return app, nil
}

func (app *App) Close() {
	if app.closed {
		return
	}

	app.closed = true

	close(app.quit)

	app.listener.Close()

	if app.httpListener != nil {
		app.httpListener.Close()
	}

	app.m.Close()

	if app.access != nil {
		app.access.Close()
	}

	app.ldb.Close()
}

func (app *App) Run() {
	if len(app.cfg.SlaveOf) > 0 {
		app.slaveof(app.cfg.SlaveOf)
	}

	go app.httpServe()

	for !app.closed {
		conn, err := app.listener.Accept()
		if err != nil {
			continue
		}

		newClientRESP(conn, app)
	}
}

func (app *App) httpServe() {
	if app.httpListener == nil {
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		newClientHTTP(app, w, r)
	})

	svr := http.Server{Handler: mux}
	svr.Serve(app.httpListener)
}

func (app *App) Ledis() *ledis.Ledis {
	return app.ldb
}
