package main

import (
	"encoding/json"
	"flag"
	"github.com/siddontang/copier"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/leveldb"
	"io/ioutil"
	"path"
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

	dbPath := path.Join(cfg.DataDir, "data")

	dbCfg := new(leveldb.Config)
	copier.Copy(dbCfg, &cfg.DB)
	dbCfg.Path = dbPath

	if err = leveldb.Repair(dbCfg); err != nil {
		println("repair error: ", err.Error())
	}
}
