package server

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/siddontang/ledisdb/ledis"
	"io"
)

//ledis <-> lua type conversion, same as http://redis.io/commands/eval

type luaClient struct {
	l *lua.State
}

type luaWriter struct {
	l *lua.State
}

func (w *luaWriter) writeError(err error) {
	w.l.NewTable()
	top := w.l.GetTop()

	w.l.PushString("err")
	w.l.PushString(err.Error())
	w.l.SetTable(top)
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
		w.l.PushString(ledis.String(b))
	}
}

func (w *luaWriter) writeArray(lst []interface{}) {
	if lst == nil {
		w.l.PushBoolean(false)
		return
	}

	base := w.l.GetTop()
	defer func() {
		if e := recover(); e != nil {
			w.l.SetTop(base)
			w.writeError(fmt.Errorf("%v", e))
		}
	}()

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
	top := w.l.GetTop()
	for i, v := range lst {
		w.l.PushInteger(int64(i) + 1)
		w.l.PushString(ledis.String(v))
		w.l.SetTable(top)
	}
}

func (w *luaWriter) writeFVPairArray(lst []ledis.FVPair) {
	if lst == nil {
		w.l.PushBoolean(false)
		return
	}

	w.l.CreateTable(len(lst)*2, 0)
	top := w.l.GetTop()
	for i, v := range lst {
		w.l.PushInteger(int64(2*i) + 1)
		w.l.PushString(ledis.String(v.Field))
		w.l.SetTable(top)

		w.l.PushInteger(int64(2*i) + 2)
		w.l.PushString(ledis.String(v.Value))
		w.l.SetTable(top)
	}
}

func (w *luaWriter) writeScorePairArray(lst []ledis.ScorePair, withScores bool) {
	if lst == nil {
		w.l.PushBoolean(false)
		return
	}

	if withScores {
		w.l.CreateTable(len(lst)*2, 0)
		top := w.l.GetTop()
		for i, v := range lst {
			w.l.PushInteger(int64(2*i) + 1)
			w.l.PushString(ledis.String(v.Member))
			w.l.SetTable(top)

			w.l.PushInteger(int64(2*i) + 2)
			w.l.PushString(ledis.String(ledis.StrPutInt64(v.Score)))
			w.l.SetTable(top)
		}
	} else {
		w.l.CreateTable(len(lst), 0)
		top := w.l.GetTop()
		for i, v := range lst {
			w.l.PushInteger(int64(i) + 1)
			w.l.PushString(ledis.String(v.Member))
			w.l.SetTable(top)
		}
	}
}

func (w *luaWriter) writeBulkFrom(n int64, rb io.Reader) {
	w.writeError(fmt.Errorf("unsupport"))
}

func (w *luaWriter) flush() {

}
