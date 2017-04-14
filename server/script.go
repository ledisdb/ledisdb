package server

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/siddontang/go/hack"
	"github.com/siddontang/go/num"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/yuin/gopher-lua"

	luajson "github.com/glendc/gopher-json"
)

//ledis <-> lua type conversion, same as http://redis.io/commands/eval

type luaWriter struct {
	l *lua.LState
}

func (w *luaWriter) writeError(err error) {
	panic(err)
}

func (w *luaWriter) writeStatus(status string) {
	table := w.l.NewTable()

	table.Append(lua.LString("ok"))
	table.Append(lua.LString(status))

	w.l.Push(table)
}

func (w *luaWriter) writeInteger(n int64) {
	w.l.Push(w.toLuaInteger(n))
}

func (w *luaWriter) writeBulk(b []byte) {
	w.l.Push(w.toLuaBulk(b))
}

func (w *luaWriter) writeArray(lst []interface{}) {
	w.l.Push(w.toLuaArray(lst))
}

func (w *luaWriter) writeSliceArray(lst [][]byte) {
	w.l.Push(w.toLuaSliceArray(lst))
}

func (w *luaWriter) writeFVPairArray(lst []ledis.FVPair) {
	if lst == nil {
		w.l.Push(lua.LFalse)
		return
	}

	table := w.l.CreateTable(len(lst)*2, 0)

	for _, v := range lst {
		table.Append(lua.LString(hack.String(v.Field)))
		table.Append(lua.LString(hack.String(v.Value)))
	}

	w.l.Push(table)
}

func (w *luaWriter) writeScorePairArray(lst []ledis.ScorePair, withScores bool) {
	if lst == nil {
		w.l.Push(lua.LFalse)
		return
	}

	var table *lua.LTable

	if withScores {
		table = w.l.CreateTable(len(lst)*2, 0)

		for _, v := range lst {
			table.Append(lua.LString(hack.String(v.Member)))
			table.Append(lua.LString(num.FormatInt64ToSlice(v.Score)))
		}
	} else {
		table = w.l.CreateTable(len(lst), 0)

		for _, v := range lst {
			table.Append(lua.LString(hack.String(v.Member)))
		}
	}

	w.l.Push(table)
}

func (w *luaWriter) writeBulkFrom(n int64, rb io.Reader) {
	w.writeError(errors.New("unsupport"))
}

func (w *luaWriter) flush() {
}

func (w *luaWriter) toLuaInteger(n int64) lua.LValue {
	return lua.LNumber(n)
}

func (w *luaWriter) toLuaBulk(b []byte) lua.LValue {
	if b == nil {
		return lua.LFalse
	}

	return lua.LString(hack.String(b))
}

func (w *luaWriter) toLuaSliceArray(lst [][]byte) lua.LValue {
	if lst == nil {
		return lua.LFalse
	}

	table := w.l.CreateTable(len(lst), 0)

	for _, v := range lst {
		if v == nil {
			table.Append(lua.LFalse)
		} else {
			table.Append(lua.LString((hack.String(v))))
		}
	}

	return table
}

func (w *luaWriter) toLuaArray(lst []interface{}) lua.LValue {
	if lst == nil {
		return lua.LFalse
	}

	table := w.l.CreateTable(len(lst), 0)

	for i := range lst {
		switch v := lst[i].(type) {
		case []interface{}:
			table.Append(w.toLuaArray(v))
		case [][]byte:
			table.Append(w.toLuaSliceArray(v))
		case []byte:
			table.Append(w.toLuaBulk(v))
		case nil:
			table.Append(w.toLuaBulk(nil))
		case int64:
			table.Append(w.toLuaInteger(v))
		default:
			panic("invalid array type")
		}
	}

	return table
}

type script struct {
	sync.Mutex

	app *App
	l   *lua.LState
	c   *client

	chunks map[string]struct{}
}

func (app *App) openScript() {
	s := new(script)
	s.app = app

	s.chunks = make(map[string]struct{})

	app.script = s

	l := lua.NewState()

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.MathLibName, lua.OpenMath},
		{lua.StringLibName, lua.OpenString},
		{lua.TabLibName, lua.OpenTable},
		{luajson.CJsonLibName, luajson.OpenCJSON},
		// TODO (gopher-lua): support libs:
		// + CMsgpackLib?! (which funcs?)
		// + StructLib?! (which funcs?)
	} {
		l.Push(l.NewFunction(pair.f))
		l.Push(lua.LString(pair.n))
		l.Call(1, 0)
	}

	l.Register("error", luaErrorHandler)

	s.l = l
	s.c = newClient(app)
	s.c.db = nil

	w := new(luaWriter)
	w.l = l
	s.c.resp = w

	setLuaDBGlobalVar(l, "ledis")
	setLuaDBGlobalVar(l, "redis")

	setMapState(l, s)
}

func (app *App) closeScript() {
	app.script.l.Close()
	delMapState(app.script.l)
	app.script = nil
}

var mapState = map[*lua.LState]*script{}
var stateLock sync.Mutex

func setLuaDBGlobalVar(l *lua.LState, name string) {
	mt := l.NewTypeMetatable(name)
	l.SetGlobal(name, mt)
	// static attributes
	l.SetField(mt, "call", l.NewFunction(luaCall))
	l.SetField(mt, "pcall", l.NewFunction(luaPCall))
	l.SetField(mt, "sha1hex", l.NewFunction(luaSha1Hex))
	l.SetField(mt, "error_reply", l.NewFunction(luaErrorReply))
	l.SetField(mt, "status_reply", l.NewFunction(luaStatusReply))
}

func setMapState(l *lua.LState, s *script) {
	stateLock.Lock()
	defer stateLock.Unlock()

	mapState[l] = s
}

func getMapState(l *lua.LState) *script {
	stateLock.Lock()
	defer stateLock.Unlock()

	return mapState[l]
}

func delMapState(l *lua.LState) {
	stateLock.Lock()
	defer stateLock.Unlock()

	delete(mapState, l)
}

func luaErrorHandler(l *lua.LState) int {
	msg := l.ToString(1)
	panic(errors.New(msg))
}

func luaCall(l *lua.LState) int {
	return luaCallGenericCommand(l)
}

func luaPCall(l *lua.LState) (n int) {
	defer func() {
		if e := recover(); e != nil {
			luaPushError(l, fmt.Sprintf("%v", e))
			n = 1
		}
		return
	}()
	return luaCallGenericCommand(l)
}

func luaErrorReply(l *lua.LState) int {
	return luaReturnSingleFieldTable(l, "err")
}

func luaStatusReply(l *lua.LState) int {
	return luaReturnSingleFieldTable(l, "ok")
}

func luaReturnSingleFieldTable(l *lua.LState, filed string) int {
	if l.GetTop() != 1 || l.Get(-1).Type() != lua.LTString {
		luaPushError(l, "wrong number or type of arguments")
		return 1
	}

	table := l.NewTable()
	table.Append(lua.LString(filed))
	l.Push(table)
	return 1
}

func luaSha1Hex(l *lua.LState) int {
	if argc := l.GetTop(); argc != 1 {
		luaPushError(l, "wrong number of arguments")
		return 1
	}

	s := l.ToString(1)
	s = hex.EncodeToString(hack.Slice(s))

	l.Push(lua.LString(s))
	return 1
}

func luaPushError(l *lua.LState, msg string) {
	table := l.NewTable()
	table.Append(lua.LString("err"))
	table.Append(lua.LString(msg))
	l.Push(table)
}

func luaCallGenericCommand(l *lua.LState) int {
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
		switch l.Get(i).Type() {
		case lua.LTNumber:
			c.args[i-2] = []byte(fmt.Sprintf("%.17g", l.ToNumber(i)))
		case lua.LTString:
			c.args[i-2] = []byte(l.ToString(i))
		default:
			panic("Lua ledis() command arguments must be strings or integers")
		}
	}

	c.perform()

	return 1
}

func luaSetGlobalArray(l *lua.LState, name string, ay [][]byte) {
	table := l.NewTable()

	for i := 0; i < len(ay); i++ {
		table.Append(lua.LString(hack.String(ay[i])))
	}

	l.SetGlobal(name, table)
}

func luaReplyToLedisReply(l *lua.LState) interface{} {
	return luaValueToLedisValue(l.Get(-1))
}

func luaValueToLedisValue(v lua.LValue) interface{} {
	switch top := v.(type) {
	case lua.LString:
		return hack.Slice(top.String())
	case lua.LBool:
		if top == lua.LTrue {
			return int64(1)
		}
		return nil
	case lua.LNumber:
		return int64(top)
	case *lua.LTable:
		// flatten all key, values, for easier access later
		flatTable := make([]lua.LValue, 0)
		var err error
		top.ForEach(func(key, value lua.LValue) {
			if err != nil {
				return
			}
			if key.Type() == lua.LTString {
				err = fmt.Errorf("only array-tables are supported: %q", top.String())
				return
			}
			flatTable = append(flatTable, key, value)
		})
		if err != nil {
			return err
		}

		length := len(flatTable)
		if length == 0 {
			return nil
		}

		if length <= 4 {
			// ok => status Reply
			// err => error Reply
			if flatTable[1].Type() == lua.LTString {
				switch strings.ToLower(flatTable[1].String()) {
				case "ok":
					if length == 4 {
						return flatTable[3].String()
					}
					return "ok"
				case "err":
					if length == 4 {
						return errors.New(flatTable[3].String())
					}
					return errors.New("err")
				default:
				}
			}
		}

		ay := make([]interface{}, 0)
		for i := 0; i < length; i += 2 {
			// cut at first nil value
			value := flatTable[i+1]
			if value.Type() == lua.LTNil {
				break
			}

			ay = append(ay, luaValueToLedisValue(value))
		}

		return ay

	default:
		return nil
	}
}
