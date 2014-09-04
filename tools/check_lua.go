// +build ignore

package main

import "github.com/siddontang/golua/lua"

func main() {
	L := lua.NewState()
	L.Close()
}
