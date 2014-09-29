// +build lua

package server

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/siddontang/go/hack"

	"github.com/siddontang/ledisdb/lua"
	"strconv"
	"strings"
)

func parseEvalArgs(l *lua.State, c *client) error {
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

func evalGenericCommand(c *client, evalSha1 bool) error {
	m, err := c.db.Multi()
	if err != nil {
		return err
	}

	s := c.app.s
	luaClient := s.c
	l := s.l

	s.Lock()

	base := l.GetTop()

	defer func() {
		l.SetTop(base)
		luaClient.db = nil
		luaClient.script = nil

		s.Unlock()

		m.Close()
	}()

	luaClient.db = m.DB
	luaClient.script = m
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

	l.GetGlobal(key)

	if l.IsNil(-1) {
		l.Pop(1)

		if evalSha1 {
			return fmt.Errorf("missing %s script", key)
		}

		if r := l.LoadString(hack.String(c.args[0])); r != 0 {
			err := fmt.Errorf("%s", l.ToString(-1))
			l.Pop(1)
			return err
		} else {
			l.PushValue(-1)
			l.SetGlobal(key)

			s.chunks[key] = struct{}{}
		}
	}

	if err := l.Call(0, lua.LUA_MULTRET); err != nil {
		return err
	} else {
		r := luaReplyToLedisReply(l)
		m.Close()

		if v, ok := r.(error); ok {
			return v
		}

		writeValue(c.resp, r)
	}

	return nil
}

func evalCommand(c *client) error {
	return evalGenericCommand(c, false)
}

func evalshaCommand(c *client) error {
	return evalGenericCommand(c, true)
}

func scriptCommand(c *client) error {
	s := c.app.s
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

	return nil
}

func scriptLoadCommand(c *client) error {
	s := c.app.s
	l := s.l

	if len(c.args) != 2 {
		return ErrCmdParams
	}

	h := sha1.Sum(c.args[1])
	key := hex.EncodeToString(h[0:20])

	if r := l.LoadString(hack.String(c.args[1])); r != 0 {
		err := fmt.Errorf("%s", l.ToString(-1))
		l.Pop(1)
		return err
	} else {
		l.PushValue(-1)
		l.SetGlobal(key)

		s.chunks[key] = struct{}{}
	}

	c.resp.writeBulk(hack.Slice(key))
	return nil
}

func scriptExistsCommand(c *client) error {
	s := c.app.s

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
	s := c.app.s
	l := s.l

	if len(c.args) != 1 {
		return ErrCmdParams
	}

	for n, _ := range s.chunks {
		l.PushNil()
		l.SetGlobal(n)
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
