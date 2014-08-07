package main

import (
	"flag"
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

var configPath = flag.String("config", "", "ledisdb config file")
var dumpPath = flag.String("dump_file", "", "ledisdb dump file")

func main() {
	flag.Parse()

	if len(*configPath) == 0 {
		println("need ledis config file")
		return
	}

	cfg, err := config.NewConfigWithFile(*configPath)
	if err != nil {
		println(err.Error())
		return
	}

	if len(*dumpPath) == 0 {
		println("need dump file")
		return
	}

	if len(cfg.DataDir) == 0 {
		println("must set data dir")
		return
	}

	ldb, err := ledis.Open(cfg)
	if err != nil {
		println("ledis open error ", err.Error())
		return
	}

	err = loadDump(cfg, ldb)
	ldb.Close()

	if err != nil {
		println(err.Error())
		return
	}

	println("Load OK")
}

func loadDump(cfg *config.Config, ldb *ledis.Ledis) error {
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
		format := "MASTER_LOG_FILE='binlog.%07d', MASTER_LOG_POS=%d;\n"
		fmt.Printf(format, head.LogFileIndex, head.LogPos)
	}

	return nil
}
