// +build lua

package server

import (
	"fmt"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/lua"

	"testing"
)

var testLuaWriter = &luaWriter{}

func testLuaWriteError(l *lua.State) int {
	testLuaWriter.writeError(fmt.Errorf("test error"))
	return 1
}

func testLuaWriteArray(l *lua.State) int {
	ay := make([]interface{}, 2)
	ay[0] = []byte("1")
	b := make([]interface{}, 2)
	b[0] = int64(10)
	b[1] = []byte("11")

	ay[1] = b

	testLuaWriter.writeArray(ay)

	return 1
}

func TestLuaWriter(t *testing.T) {
	l := lua.NewState()

	l.OpenBase()

	testLuaWriter.l = l

	l.Register("WriteError", testLuaWriteError)

	str := `
        WriteError() 
        `

	err := l.DoString(str)

	if err == nil {
		t.Fatal("must error")
	}

	l.Register("WriteArray", testLuaWriteArray)

	str = `
        local  a = WriteArray()

        if #a ~= 2 then
            error("len a must 2")
        elseif a[1] ~= "1" then 
            error("a[1] must 1")
        elseif #a[2] ~= 2 then 
            error("len a[2] must 2")
        elseif a[2][1] ~= 10 then
            error("a[2][1] must 10")
        elseif a[2][2] ~= "11" then
            error("a[2][2] must 11") 
        end
        `

	err = l.DoString(str)
	if err != nil {
		t.Fatal(err)
	}

	l.Close()
}

var testScript1 = `
    return {1,2,3} 
`

var testScript2 = `
    return ledis.call("ping")
`

var testScript3 = `
    ledis.call("set", 1, "a")

    local a = ledis.call("get", 1)
    if type(a) ~= "string" then
        error("must string")
    elseif a ~= "a" then 
        error("must a")
    end
`

var testScript4 = `
    ledis.call("select", 2)
    ledis.call("set", 2, "b")
`

func TestLuaCall(t *testing.T) {
	cfg := config.NewConfigDefault()
	cfg.Addr = ":11188"
	cfg.DataDir = "/tmp/testscript"
	cfg.DBName = "memory"

	app, e := NewApp(cfg)
	if e != nil {
		t.Fatal(e)
	}
	go app.Run()

	defer app.Close()

	db, _ := app.ldb.Select(0)
	m, _ := db.Multi()
	defer m.Close()

	luaClient := app.s.c
	luaClient.db = m.DB
	luaClient.script = m

	l := app.s.l

	err := app.s.l.DoString(testScript1)
	if err != nil {
		t.Fatal(err)
	}

	v := luaReplyToLedisReply(l)
	if vv, ok := v.([]interface{}); ok {
		if len(vv) != 3 {
			t.Fatal(len(vv))
		}
	} else {
		t.Fatal(fmt.Sprintf("%v %T", v, v))
	}

	err = app.s.l.DoString(testScript2)
	if err != nil {
		t.Fatal(err)
	}

	v = luaReplyToLedisReply(l)
	if vv := v.(string); vv != "PONG" {
		t.Fatal(fmt.Sprintf("%v %T", v, v))
	}

	err = app.s.l.DoString(testScript3)
	if err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get([]byte("1")); err != nil {
		t.Fatal(err)
	} else if string(v) != "a" {
		t.Fatal(string(v))
	}

	err = app.s.l.DoString(testScript4)
	if err != nil {
		t.Fatal(err)
	}

	if luaClient.db.Index() != 2 {
		t.Fatal(luaClient.db.Index())
	}

	db2, _ := app.ldb.Select(2)
	if v, err := db2.Get([]byte("2")); err != nil {
		t.Fatal(err)
	} else if string(v) != "b" {
		t.Fatal(string(v))
	}

	luaClient.db = nil
}
