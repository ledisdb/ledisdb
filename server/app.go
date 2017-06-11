package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"crypto/tls"
	"github.com/siddontang/goredis"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
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

	info *info

	script *script

	// handle slaves
	slock        sync.Mutex
	slaves       map[string]*client
	slaveSyncAck chan uint64

	snap *snapshotStore

	connWait sync.WaitGroup

	rcm sync.Mutex
	rcs map[*respClient]struct{}

	migrateM          sync.Mutex
	migrateClients    map[string]*goredis.Client
	migrateKeyLockers map[string]*migrateKeyLocker
}

func netType(s string) string {
	if strings.Contains(s, "/") {
		return "unix"
	}

	return "tcp"
}

func tlsConfig(c *config.TLS) (*tls.Config, error) {
	crt, err := tls.LoadX509KeyPair(c.Certificate, c.Key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{
			crt,
		},
	}, nil
}

func listen(netType, laddr string, tlsCfg *tls.Config) (net.Listener, error) {
	if tlsCfg != nil {
		return tls.Listen(netType, laddr, tlsCfg)
	}

	return net.Listen(netType, laddr)
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

	app.slaves = make(map[string]*client)
	app.slaveSyncAck = make(chan uint64)

	app.rcs = make(map[*respClient]struct{})

	app.migrateClients = make(map[string]*goredis.Client)
	app.newMigrateKeyLockers()

	var err error

	if app.info, err = newInfo(app); err != nil {
		return nil, err
	}

	var tlsCfg *tls.Config
	if cfg.TLS.Enabled {
		tlsCfg, err = tlsConfig(&cfg.TLS)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Addr != "" {
		addrNetType := netType(cfg.Addr)

		if app.listener, err = listen(addrNetType, cfg.Addr, tlsCfg); err != nil {
			return nil, err
		}

		if addrNetType == "unix" && len(cfg.AddrUnixSocketPerm) > 0 {
			var perm int64
			if perm, err = strconv.ParseInt(cfg.AddrUnixSocketPerm, 8, 32); err != nil {
				return nil, err
			}
			if err = os.Chmod(cfg.Addr, os.FileMode(perm)); err != nil {
				return nil, err
			}
		}
	} else {
		app.listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			if app.listener, err = net.Listen("tcp6", "[::1]:0"); err != nil {
				return nil, fmt.Errorf("app: failed to listen on a port: %v", err)
			}
		}
	}

	if len(cfg.HttpAddr) > 0 {
		if app.httpListener, err = listen(netType(cfg.HttpAddr), cfg.HttpAddr, tlsCfg); err != nil {
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

	if app.snap, err = newSnapshotStore(cfg); err != nil {
		return nil, err
	}

	if len(app.cfg.SlaveOf) > 0 {
		//slave must readonly
		app.cfg.Readonly = true
	}

	if app.ldb, err = ledis.Open(cfg); err != nil {
		return nil, err
	}

	app.m = newMaster(app)

	app.openScript()

	app.ldb.AddNewLogEventHandler(app.publishNewLog)

	return app, nil
}

func (app *App) Close() {
	if app.closed {
		return
	}

	app.closed = true

	close(app.quit)

	app.listener.Close()

	//close all migrate connections
	app.migrateM.Lock()
	for k, c := range app.migrateClients {
		c.Close()
		delete(app.migrateClients, k)
	}
	app.migrateM.Unlock()

	if app.httpListener != nil {
		app.httpListener.Close()
	}

	app.closeAllRespClients()

	//wait all connection closed
	app.connWait.Wait()

	app.closeScript()

	app.m.Lock()
	app.m.Close()
	app.m.Unlock()

	app.snap.Close()

	if app.access != nil {
		app.access.Close()
	}

	app.ldb.Close()
}

func (app *App) Run() {
	if len(app.cfg.SlaveOf) > 0 {
		app.slaveof(app.cfg.SlaveOf, false, app.cfg.Readonly)
	}

	go app.httpServe()

	for {
		select {
		case <-app.quit:
			return
		default:
			conn, err := app.listener.Accept()
			if err != nil {
				continue
			}

			newClientRESP(conn, app)
		}
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

func (app *App) Address() string {
	return app.listener.Addr().String()
}
