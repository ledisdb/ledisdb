package main

import (
	"encoding/json"
	"flag"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/server"
	"io/ioutil"
	"path"
)

var configPath = flag.String("config", "/etc/ledis.json", "ledisdb config file")
var dumpPath = flag.String("dump_file", "", "ledisdb dump file")
var masterAddr = flag.String("master_addr", "",
	"master addr to set where dump file comes from, if not set correctly, next slaveof may cause fullsync")

func main() {
	flag.Parse()

	if len(*configPath) == 0 {
		println("need ledis config file")
		return
	}

	data, err := ioutil.ReadFile(*configPath)
	if err != nil {
		println(err.Error())
		return
	}

	if len(*dumpPath) == 0 {
		println("need dump file")
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

	ldb, err := ledis.Open(&cfg)
	if err != nil {
		println("ledis open error ", err.Error())
		return
	}

	err = loadDump(&cfg, ldb)
	ldb.Close()

	if err != nil {
		println(err.Error())
		return
	}

	println("Load OK")
}

func loadDump(cfg *ledis.Config, ldb *ledis.Ledis) error {
	var err error
	if err = ldb.FlushAll(); err != nil {
		return err
	}

	var head *ledis.MasterInfo
	head, err = ldb.LoadDumpFile(*dumpPath)

	if err != nil {
		return err
	}

	info := new(server.MasterInfo)

	info.Addr = *masterAddr
	info.LogFileIndex = head.LogFileIndex
	info.LogPos = head.LogPos

	infoFile := path.Join(cfg.DataDir, "master.info")

	return info.Save(infoFile)
}
