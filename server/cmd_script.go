package server

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/siddontang/go/hack"

	"strconv"
	"strings"

	"github.com/yuin/gopher-lua"
)

func parseEvalArgs(l *lua.LState, c *client) error {
	args := c.args
	if len(args) < 2 {
		return ErrCmdParams
	}

	args = args[1:]

	n, err := strconv.Atoi(hack.String(args[0]))
	if err != nil {
		return err
	}

	if n > len(args)-1 {
		return ErrCmdParams
	}

	luaSetGlobalArray(l, "KEYS", args[1:n+1])
	luaSetGlobalArray(l, "ARGV", args[n+1:])

	return nil
}

func evalGenericCommand(c *client, evalSha1 bool) (err error) {
	s := c.app.script
	luaClient := s.c
	l := s.l

	s.Lock()

	defer func() {
		luaClient.db = nil
		// luaClient.script = nil

		s.Unlock()
	}()

	luaClient.db = c.db
	// luaClient.script = m
	luaClient.remoteAddr = c.remoteAddr

	if err := parseEvalArgs(l, c); err != nil {
		return err
	}

	var key string
	if !evalSha1 {
		h := sha1.Sum(c.args[0])
		key = hex.EncodeToString(h[0:20])
	} else {
		key = strings.ToLower(hack.String(c.args[0]))
	}

	global := l.GetGlobal(key)

	if global.Type() == lua.LTNil {
		if evalSha1 {
			return errors.New("NOSCRIPT no matching script, please use EVAL")
		}

		val, err := l.LoadString(hack.String(c.args[0]))
		if err != nil {
			return err
		}

		l.SetGlobal(key, val)
		s.chunks[key] = struct{}{}
		global = val
	}

	l.Push(global)

	// catch any uncaught panic
	// this happens for example when the user,
	// makes a mistake using `ledis.call`
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	l.Call(0, lua.MultRet)

	r := luaReplyToLedisReply(l)
	if v, ok := r.(error); ok {
		return v
	}

	writeValue(c.resp, r)
	return nil
}

func evalCommand(c *client) error {
	return evalGenericCommand(c, false)
}

func evalshaCommand(c *client) error {
	return evalGenericCommand(c, true)
}

func scriptCommand(c *client) error {
	s := c.app.script
	l := s.l

	s.Lock()

	base := l.GetTop()

	defer func() {
		l.SetTop(base)
		s.Unlock()
	}()

	args := c.args

	if len(args) < 1 {
		return ErrCmdParams
	}

	switch strings.ToLower(hack.String(args[0])) {
	case "load":
		return scriptLoadCommand(c)
	case "exists":
		return scriptExistsCommand(c)
	case "flush":
		return scriptFlushCommand(c)
	default:
		return fmt.Errorf("invalid script %s", args[0])
	}
}

func scriptLoadCommand(c *client) error {
	s := c.app.script
	l := s.l

	if len(c.args) != 2 {
		return ErrCmdParams
	}

	h := sha1.Sum(c.args[1])
	key := hex.EncodeToString(h[0:20])

	val, err := l.LoadString(hack.String(c.args[1]))
	if err != nil {
		return err
	}
	l.Push(val)

	l.SetGlobal(key, val)
	s.chunks[key] = struct{}{}

	c.resp.writeBulk(hack.Slice(key))
	return nil
}

func scriptExistsCommand(c *client) error {
	s := c.app.script

	if len(c.args) < 2 {
		return ErrCmdParams
	}

	ay := make([]interface{}, len(c.args[1:]))
	for i, n := range c.args[1:] {
		if _, ok := s.chunks[hack.String(n)]; ok {
			ay[i] = int64(1)
		} else {
			ay[i] = int64(0)
		}
	}

	c.resp.writeArray(ay)
	return nil
}

func scriptFlushCommand(c *client) error {
	s := c.app.script
	l := s.l

	if len(c.args) != 1 {
		return ErrCmdParams
	}

	for n := range s.chunks {
		l.SetGlobal(n, lua.LNil)
	}

	s.chunks = map[string]struct{}{}

	c.resp.writeStatus(OK)

	return nil
}

func init() {
	register("eval", evalCommand)
	register("evalsha", evalshaCommand)
	register("script", scriptCommand)
}
