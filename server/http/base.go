package http

import (
	"errors"
	"fmt"
	"github.com/siddontang/ledisdb/ledis"
	"strings"
)

const ERR_ARGUMENT_FORMAT = "ERR wrong number of arguments for '%s' command"

var ErrValue = errors.New("ERR value is not an integer or out of range")

type commondFunc func(*ledis.DB, ...string) (interface{}, error)

var regCmds = map[string]commondFunc{}

func register(name string, f commondFunc) {
	if _, ok := regCmds[strings.ToLower(name)]; ok {
		panic(fmt.Sprintf("%s has been registered", name))
	}
	regCmds[name] = f
}
func lookup(name string) commondFunc {
	return regCmds[strings.ToLower(name)]
}
