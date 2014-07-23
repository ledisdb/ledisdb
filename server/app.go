package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"net"
	"path"
	"strings"
)

type App struct {
	cfg *Config

	listener net.Listener

	ldb *ledis.Ledis

	closed bool

	quit chan struct{}

	access *accessLog

	//for slave replication
	m *master
}

func NewApp(cfg *Config) (*App, error) {
	if len(cfg.DataDir) == 0 {
		return nil, fmt.Errorf("must set data_dir first")
	}

	app := new(App)

	app.quit = make(chan struct{})

	app.closed = false

	app.cfg = cfg

	var err error

	if strings.Contains(cfg.Addr, "/") {
		app.listener, err = net.Listen("unix", cfg.Addr)
	} else {
		app.listener, err = net.Listen("tcp", cfg.Addr)
	}

	if err != nil {
		return nil, err
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

	if app.ldb, err = ledis.Open(cfg.NewLedisConfig()); err != nil {
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

	for !app.closed {
		conn, err := app.listener.Accept()
		if err != nil {
			continue
		}

		newClient(conn, app)
	}
}

func (app *App) Ledis() *ledis.Ledis {
	return app.ldb
}
