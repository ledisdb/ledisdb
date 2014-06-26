package main

import (
	"encoding/json"
	"flag"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/leveldb"
	"io/ioutil"
	"path"
)

var fileName = flag.String("config", "/etc/ledis.config", "ledisdb config file")

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(fileName)
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

	if len(cfg.DataDB.Path) == 0 {
		cfg.DataDB.Path = path.Join(cfg.DataDir, "data")
	}

	if err = leveldb.Repair(&cfg.DataDB); err != nil {
		println("repair error: ", err.Error())
	}
}
