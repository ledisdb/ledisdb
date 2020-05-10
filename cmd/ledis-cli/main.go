package main

import (
	"fmt"

	"github.com/ledisdb/ledisdb/cmd"
)

var (
	version  = "dev"
	buildTag string
)

func main() {
	fmt.Printf("Version %s", version)
	if len(buildTag) > 0 {
		fmt.Printf(" with tag %s", buildTag)
	}
	fmt.Println()

	cmd.Cli()
}
