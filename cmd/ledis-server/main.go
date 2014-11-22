package main

import (
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/server"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var configFile = flag.String("config", "", "ledisdb config file")
var addr = flag.String("addr", "", "ledisdb listen address")
var dataDir = flag.String("data_dir", "", "ledisdb base data dir")
var dbName = flag.String("db_name", "", "select a db to use, it will overwrite the config's db name")
var usePprof = flag.Bool("pprof", false, "enable pprof")
var pprofPort = flag.Int("pprof_port", 6060, "pprof http port")
var slaveof = flag.String("slaveof", "", "make the server a slave of another instance")
var readonly = flag.Bool("readonly", false, "set readonly mode, salve server is always readonly")
var rpl = flag.Bool("rpl", false, "enable replication or not, slave server is always enabled")
var rplSync = flag.Bool("rpl_sync", false, "enable sync replication or not")
var ttlCheck = flag.Int("ttl_check", 0, "TTL check interval")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	var cfg *config.Config
	var err error

	if len(*configFile) == 0 {
		println("no config set, using default config")
		cfg = config.NewConfigDefault()
	} else {
		cfg, err = config.NewConfigWithFile(*configFile)
	}

	if err != nil {
		println(err.Error())
		return
	}

	if len(*addr) > 0 {
		cfg.Addr = *addr
	}

	if len(*dataDir) > 0 {
		cfg.DataDir = *dataDir
	}

	if len(*dbName) > 0 {
		cfg.DBName = *dbName
	}

	if len(*slaveof) > 0 {
		cfg.SlaveOf = *slaveof
		cfg.Readonly = true
		cfg.UseReplication = true
	} else {
		cfg.Readonly = *readonly
		cfg.UseReplication = *rpl
		cfg.Replication.Sync = *rplSync
	}

	if *ttlCheck > 0 {
		cfg.TTLCheckInterval = *ttlCheck
	}

	var app *server.App
	app, err = server.NewApp(cfg)
	if err != nil {
		println(err.Error())
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	if *usePprof {
		go func() {
			log.Println(http.ListenAndServe(fmt.Sprintf(":%d", *pprofPort), nil))
		}()
	}

	go app.Run()

	<-sc

	app.Close()
}
