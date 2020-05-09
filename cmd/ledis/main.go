package main

import (
	"fmt"
	"os"

	"github.com/ledisdb/ledisdb/cmd"
)

var (
	cmds = [][]string{
		{"server", "run ledis server"},
		{"cli", "run ledis client"},
		{"repair", "repair ledis storage directory"},
		{"dump", "create a snapshort of ledis"},
		{"load", "load data from a snapshort"},
		{"benchmark", "run the benchmarks with ledis"},
	}
)

func printSubCmds() {
	for _, cmd := range cmds {
		printCmd(cmd[0], cmd[1])
	}
}

func printCmd(cmd, description string) {
	fmt.Printf("%s\t- %s\n", cmd, description)
}

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
	case "help":
		printSubCmds()
	case "server":
		fallthrough
	default:
		cmd.CmdServer()
	}
}
