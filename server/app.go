package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/replication"
	"net"
	"strings"
)

type App struct {
	cfg *Config

	listener net.Listener

	ldb *ledis.Ledis

	closed bool

	slaveMode bool

	relayLog *replication.Log

	quit chan struct{}
}

func NewApp(cfg *Config) (*App, error) {
	if len(cfg.DataDir) == 0 {
		return nil, fmt.Errorf("must set data_dir first")
	}

	if len(cfg.DB.DataDir) == 0 {
		cfg.DB.DataDir = cfg.DataDir
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

	app.slaveMode = false

	if len(app.cfg.SlaveOf) > 0 {
		app.slaveMode = true

		if app.relayLog, err = replication.NewRelayLogWithConfig(&cfg.RelayLog); err != nil {
			return nil, err
		}
	}

	if app.ldb, err = ledis.OpenWithConfig(&cfg.DB); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *App) Close() {
	if app.closed {
		return
	}

	app.closed = true

	close(app.quit)

	app.listener.Close()

	app.ldb.Close()
}

func (app *App) Run() {
	if app.slaveMode {
		app.runReplication()
	}

	for !app.closed {
		conn, err := app.listener.Accept()
		if err != nil {
			continue
		}

		newClient(conn, app)
	}
}
