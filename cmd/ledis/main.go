package main

import (
	"os"

	"github.com/ledisdb/ledisdb/cmd"
)

func main() {
	var subCmd string
	if len(os.Args) == 1 {
		subCmd = "server"
	} else {
		subCmd = os.Args[1]
	}

	switch subCmd {
	case "repair":
		cmd.CmdRepair()
	case "benchmark":
		cmd.CmdBenchmark()
	case "cli":
		cmd.CmdCli()
	case "dump":
		cmd.CmdDump()
	case "server":
		fallthrough
	default:
		cmd.CmdServer()
	}
}
