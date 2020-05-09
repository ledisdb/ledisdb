package cmd

import (
	"flag"

	"github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
)

func CmdLoad() {
	var configPath = flag.String("config", "", "ledisdb config file")
	var dumpPath = flag.String("dump_file", "", "ledisdb dump file")

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

	err = loadDump(cfg, ldb, dumpPath)
	ldb.Close()

	if err != nil {
		println(err.Error())
		return
	}

	println("Load OK")
}

func loadDump(cfg *config.Config, ldb *ledis.Ledis, dumpPath *string) error {
	var err error
	if err = ldb.FlushAll(); err != nil {
		return err
	}

	_, err = ldb.LoadDumpFile(*dumpPath)
	return err
}
