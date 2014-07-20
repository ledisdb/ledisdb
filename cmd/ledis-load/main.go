package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"io/ioutil"
)

var configPath = flag.String("config", "/etc/ledis.json", "ledisdb config file")
var dumpPath = flag.String("dump_file", "", "ledisdb dump file")

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

	//master enable binlog, here output this like mysql
	if head.LogFileIndex != 0 && head.LogPos != 0 {
		format := "-- CHANGE MASTER TO MASTER_LOG_FILE='binlog.%07d', MASTER_LOG_POS=%d;\n"
		fmt.Printf(format, head.LogFileIndex, head.LogPos)
	}

	return nil
}
