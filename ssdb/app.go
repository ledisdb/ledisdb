package ssdb

import (
	"github.com/siddontang/golib/leveldb"
	"net"
	"strings"
	"sync"
)

type App struct {
	cfg *Config

	listener net.Listener

	db *leveldb.DB

	kvMutex   sync.Mutex
	hashMutex sync.Mutex
	listMutex sync.Mutex
	zsetMutex sync.Mutex
}

func NewApp(cfg *Config) (*App, error) {
	app := new(App)

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

	return app, nil
}

func (app *App) Close() {
	app.listener.Close()

	app.db.Close()
}

func (app *App) Run() {
	for {
		conn, err := app.listener.Accept()
		if err != nil {
			continue
		}

		newClient(conn, app)
	}
}
