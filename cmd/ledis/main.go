package main

import (
	"fmt"
	"os"

	_ "net/http/pprof"

	"github.com/ledisdb/ledisdb/cmd"
)

var (
	version  = "dev"
	buildTag string

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
	fmt.Printf("Version %s", version)
	if len(buildTag) > 0 {
		fmt.Printf(" with tag %s", buildTag)
	}
	fmt.Println()
	var subCmd string
	if len(os.Args) == 1 {
		subCmd = "server"
	} else {
		subCmd = os.Args[1]
	}

	switch subCmd {
	case "repair":
		cmd.Repair()
	case "benchmark":
		cmd.Benchmark()
	case "cli":
		cmd.Cli()
	case "dump":
		cmd.Dump()
	case "help":
		printSubCmds()
	case "server":
		fallthrough
	default:
		cmd.Server()
	}
}
