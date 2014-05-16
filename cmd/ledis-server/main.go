package main

import (
	"flag"
	"github.com/siddontang/ledisdb/server"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var configFile = flag.String("config", "", "ledisdb config file")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	if len(*configFile) == 0 {
		panic("must use a config file")
	}

	cfg, err := server.NewConfigWithFile(*configFile)
	if err != nil {
		panic(err)
	}

	var app *server.App
	app, err = server.NewApp(cfg)
	if err != nil {
		panic(err)
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

	app.Run()
}
