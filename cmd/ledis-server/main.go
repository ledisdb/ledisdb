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
var dbName = flag.String("db_name", "", "select a db to use, it will overwrite the config's db name")
var usePprof = flag.Bool("pprof", false, "enable pprof")
var pprofPort = flag.Int("pprof_port", 6060, "pprof http port")

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

	if len(*dbName) > 0 {
		cfg.DBName = *dbName
	}

	var app *server.App
	app, err = server.NewApp(cfg)
	if err != nil {
		println(err.Error())
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-sc

		app.Close()
	}()

	if *usePprof {
		go func() {
			log.Println(http.ListenAndServe(fmt.Sprintf(":%d", *pprofPort), nil))
		}()
	}

	app.Run()
}
