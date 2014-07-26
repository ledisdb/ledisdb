package main

import (
	"encoding/json"
	"flag"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/store"
	"io/ioutil"
)

var fileName = flag.String("config", "/etc/ledis.json", "ledisdb config file")

func main() {
	flag.Parse()

	if len(*fileName) == 0 {
		println("need ledis config file")
		return
	}

	data, err := ioutil.ReadFile(*fileName)
	if err != nil {
		println(err.Error())
		return
	}

	var cfg ledis.Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		println(err.Error())
		return
	}

	if len(cfg.DataDir) == 0 {
		println("must set data dir")
		return
	}

	if err = store.Repair(cfg.NewDBConfig()); err != nil {
		println("repair error: ", err.Error())
	}
}
