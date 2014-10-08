// +build lua

package server

import (
	"encoding/hex"
	"fmt"
	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/siddontang/ledisdb/lua"
	"io"
	"sync"
)

//ledis <-> lua type conversion, same as http://redis.io/commands/eval

type luaWriter struct {
	l *lua.State
}

func (w *luaWriter) writeError(err error) {
	panic(err)
}

func (w *luaWriter) writeStatus(status string) {
	w.l.NewTable()
	top := w.l.GetTop()

	w.l.PushString("ok")
	w.l.PushString(status)
	w.l.SetTable(top)
}

func (w *luaWriter) writeInteger(n int64) {
	w.l.PushInteger(n)
}

func (w *luaWriter) writeBulk(b []byte) {
	if b == nil {
		w.l.PushBoolean(false)
	} else {
		w.l.PushString(hack.String(b))
	}
}

func (w *luaWriter) writeArray(lst []interface{}) {
	if lst == nil {
		w.l.PushBoolean(false)
		return
	}

	w.l.CreateTable(len(lst), 0)
	top := w.l.GetTop()

	for i, _ := range lst {
		w.l.PushInteger(int64(i) + 1)

		switch v := lst[i].(type) {
		case []interface{}:
			w.writeArray(v)
		case [][]byte:
			w.writeSliceArray(v)
		case []byte:
			w.writeBulk(v)
		case nil:
			w.writeBulk(nil)
		case int64:
			w.writeInteger(v)
		default:
			panic("invalid array type")
		}

		w.l.SetTable(top)
	}
}

func (w *luaWriter) writeSliceArray(lst [][]byte) {
	if lst == nil {
		w.l.PushBoolean(false)
		return
	}

	w.l.CreateTable(len(lst), 0)
	for i, v := range lst {
		w.l.PushString(hack.String(v))
		w.l.RawSeti(-2, i+1)
	}
}

func (w *luaWriter) writeFVPairArray(lst []ledis.FVPair) {
	if lst == nil {
		w.l.PushBoolean(false)
		return
	}

	w.l.CreateTable(len(lst)*2, 0)
	for i, v := range lst {
		w.l.PushString(hack.String(v.Field))
		w.l.RawSeti(-2, 2*i+1)

		w.l.PushString(hack.String(v.Value))
		w.l.RawSeti(-2, 2*i+2)
	}
}

func (w *luaWriter) writeScorePairArray(lst []ledis.ScorePair, withScores bool) {
	if lst == nil {
		w.l.PushBoolean(false)
		return
	}

	if withScores {
		w.l.CreateTable(len(lst)*2, 0)
		for i, v := range lst {
			w.l.PushString(hack.String(v.Member))
			w.l.RawSeti(-2, 2*i+1)

			w.l.PushString(hack.String(num.FormatInt64ToSlice(v.Score)))
			w.l.RawSeti(-2, 2*i+2)
		}
	} else {
		w.l.CreateTable(len(lst), 0)
		for i, v := range lst {
			w.l.PushString(hack.String(v.Member))
			w.l.RawSeti(-2, i+1)
		}
	}
}

func (w *luaWriter) writeBulkFrom(n int64, rb io.Reader) {
	w.writeError(fmt.Errorf("unsupport"))
}

func (w *luaWriter) flush() {

}

type script struct {
	sync.Mutex

	app *App
	l   *lua.State
	c   *client

	chunks map[string]struct{}
}

func (app *App) openScript() {
	s := new(script)
	s.app = app

	s.chunks = make(map[string]struct{})

	app.s = s

	l := lua.NewState()

	l.OpenBase()
	l.OpenLibs()
	l.OpenMath()
	l.OpenString()
	l.OpenTable()
	l.OpenPackage()

	l.OpenCJson()
	l.OpenCMsgpack()
	l.OpenStruct()

	l.Register("error", luaErrorHandler)

	s.l = l
	s.c = newClient(app)
	s.c.db = nil

	w := new(luaWriter)
	w.l = l
	s.c.resp = w

	l.NewTable()
	l.PushString("call")
	l.PushGoFunction(luaCall)
	l.SetTable(-3)

	l.PushString("pcall")
	l.PushGoFunction(luaPCall)
	l.SetTable(-3)

	l.PushString("sha1hex")
	l.PushGoFunction(luaSha1Hex)
	l.SetTable(-3)

	l.PushString("error_reply")
	l.PushGoFunction(luaErrorReply)
	l.SetTable(-3)

	l.PushString("status_reply")
	l.PushGoFunction(luaStatusReply)
	l.SetTable(-3)

	l.SetGlobal("ledis")

	setMapState(l, s)
}

func (app *App) closeScript() {
	app.s.l.Close()
	delMapState(app.s.l)
	app.s = nil
}

var mapState = map[*lua.State]*script{}
var stateLock sync.Mutex

func setMapState(l *lua.State, s *script) {
	stateLock.Lock()
	defer stateLock.Unlock()

	mapState[l] = s
}

func getMapState(l *lua.State) *script {
	stateLock.Lock()
	defer stateLock.Unlock()

	return mapState[l]
}

func delMapState(l *lua.State) {
	stateLock.Lock()
	defer stateLock.Unlock()

	delete(mapState, l)
}

func luaErrorHandler(l *lua.State) int {
	msg := l.ToString(1)
	panic(fmt.Errorf(msg))
}

func luaCall(l *lua.State) int {
	return luaCallGenericCommand(l)
}

func luaPCall(l *lua.State) (n int) {
	defer func() {
		if e := recover(); e != nil {
			luaPushError(l, fmt.Sprintf("%v", e))
			n = 1
		}
		return
	}()
	return luaCallGenericCommand(l)
}

func luaErrorReply(l *lua.State) int {
	return luaReturnSingleFieldTable(l, "err")
}

func luaStatusReply(l *lua.State) int {
	return luaReturnSingleFieldTable(l, "ok")
}

func luaReturnSingleFieldTable(l *lua.State, filed string) int {
	if l.GetTop() != 1 || l.Type(-1) != lua.LUA_TSTRING {
		luaPushError(l, "wrong number or type of arguments")
		return 1
	}

	l.NewTable()
	l.PushString(filed)
	l.PushValue(-3)
	l.SetTable(-3)
	return 1
}

func luaSha1Hex(l *lua.State) int {
	argc := l.GetTop()
	if argc != 1 {
		luaPushError(l, "wrong number of arguments")
		return 1
	}

	s := l.ToString(1)
	s = hex.EncodeToString(hack.Slice(s))

	l.PushString(s)
	return 1
}

func luaPushError(l *lua.State, msg string) {
	l.NewTable()
	l.PushString("err")
	err := l.NewError(msg)
	l.PushString(err.Error())
	l.SetTable(-3)
}

func luaCallGenericCommand(l *lua.State) int {
	s := getMapState(l)
	if s == nil {
		panic("Invalid lua call")
	} else if s.c.db == nil {
		panic("Invalid lua call, not prepared")
	}

	c := s.c

	argc := l.GetTop()
	if argc < 1 {
		panic("Please specify at least one argument for ledis.call()")
	}

	c.cmd = l.ToString(1)

	c.args = make([][]byte, argc-1)

	for i := 2; i <= argc; i++ {
		switch l.Type(i) {
		case lua.LUA_TNUMBER:
			c.args[i-2] = []byte(fmt.Sprintf("%.17g", l.ToNumber(i)))
		case lua.LUA_TSTRING:
			c.args[i-2] = []byte(l.ToString(i))
		default:
			panic("Lua ledis() command arguments must be strings or integers")
		}
	}

	c.perform()

	return 1
}

func luaSetGlobalArray(l *lua.State, name string, ay [][]byte) {
	l.NewTable()

	for i := 0; i < len(ay); i++ {
		l.PushString(hack.String(ay[i]))
		l.RawSeti(-2, i+1)
	}

	l.SetGlobal(name)
}

func luaReplyToLedisReply(l *lua.State) interface{} {
	base := l.GetTop()
	defer func() {
		l.SetTop(base - 1)
	}()

	switch l.Type(-1) {
	case lua.LUA_TSTRING:
		return hack.Slice(l.ToString(-1))
	case lua.LUA_TBOOLEAN:
		if l.ToBoolean(-1) {
			return int64(1)
		} else {
			return nil
		}
	case lua.LUA_TNUMBER:
		return int64(l.ToInteger(-1))
	case lua.LUA_TTABLE:
		l.PushString("err")
		l.GetTable(-2)
		if l.Type(-1) == lua.LUA_TSTRING {
			return fmt.Errorf("%s", l.ToString(-1))
		}

		l.Pop(1)
		l.PushString("ok")
		l.GetTable(-2)
		if l.Type(-1) == lua.LUA_TSTRING {
			return l.ToString(-1)
		} else {
			l.Pop(1)

			ay := make([]interface{}, 0)

			for i := 1; ; i++ {
				l.PushInteger(int64(i))
				l.GetTable(-2)
				if l.Type(-1) == lua.LUA_TNIL {
					l.Pop(1)
					break
				}

				ay = append(ay, luaReplyToLedisReply(l))
			}
			return ay

		}
	default:
		return nil
	}
}
