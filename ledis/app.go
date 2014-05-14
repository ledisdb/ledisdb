package ledis

import (
	"github.com/siddontang/go-leveldb/leveldb"
	"net"
	"strings"
)

type App struct {
	cfg *Config

	listener net.Listener

	db *leveldb.DB

	kvTx   *tx
	listTx *tx
	hashTx *tx
	zsetTx *tx

	closed bool
}

func NewApp(cfg *Config) (*App, error) {
	app := new(App)

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

	app.db, err = leveldb.OpenWithConfig(&cfg.DB)
	if err != nil {
		return nil, err
	}

	app.kvTx = app.newTx()
	app.listTx = app.newTx()
	app.hashTx = app.newTx()
	app.zsetTx = app.newTx()

	return app, nil
}

func (app *App) Close() {
	if app.closed {
		return
	}

	app.listener.Close()

	app.db.Close()

	app.closed = true
}

func (app *App) Run() {
	for !app.closed {
		conn, err := app.listener.Accept()
		if err != nil {
			continue
		}

		newClient(conn, app)
	}
}
