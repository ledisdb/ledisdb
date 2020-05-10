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
		{"server", "Run ledis server"},
		{"cli", "Run ledis client"},
		{"repair", "Repair ledis storage directory"},
		{"dump", "Create a snapshort of ledis"},
		{"load", "Load data from a snapshort"},
		{"benchmark", "Run the benchmarks with ledis"},
		{"repair-ttl", "Repair a very serious bug for key expiration and TTL before v0.4"},
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
	case "repair-ttl":
		cmd.RepairTTL()
	case "help":
		printSubCmds()
	case "server":
		fallthrough
	default:
		cmd.Server()
	}
}
