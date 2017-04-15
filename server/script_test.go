package server

import (
	"fmt"

	"github.com/siddontang/ledisdb/config"
	"github.com/yuin/gopher-lua"

	"testing"
)

var testLuaWriter = &luaWriter{}

func testLuaWriteError(l *lua.LState) int {
	testLuaWriter.writeError(fmt.Errorf("test error"))
	return 1
}

func testLuaWriteArray(l *lua.LState) int {
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
	l := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer l.Close()
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
	} {
		if err := l.CallByParam(lua.P{
			Fn:      l.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}

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

var testScript5 = `
    return ledis.call("PING")
`

var testScript6 = `
	ledis.call('hmset', 'zzz', 1, 2, 5, 42)
	local a = ledis.call('hmget', 'zzz', 1, 2, 5, 42)
	for i = 1, 5 do
		a[i] = type(a[i])
	end
	return a
`

var testScript7 = `
	redis.call('HSET', 'carlos', 1, 1)
	redis.call('HSET', 'carlos', 2, 2)
	redis.call('HSET', 'carlos', 3, 3)
	redis.call('HSET', 'carlos', 9, 36)

	local sum = 0
	local matches = redis.call('HKEYS', 'carlos')

	for _,key in ipairs(matches) do
		local val = redis.call('HGET', 'carlos', key)
		sum = sum + tonumber(val)
	end

	return sum
`

var testScript8 = `
	local raw = cjson.encode({a=9, b=11, c=20, d=2})
	local obj = cjson.decode(raw)
	local sum = 0

	for _,val in pairs(obj) do
		sum = sum + val
	end

	if cjson.decode("True") then
		sum = sum / 2
	end

	if cjson.encode(cjson.encode("foo")) == "foo" then
		sum = sum - 2
	end

	local arr = cjson.decode("[1,2,5,5,10]")
	for _,val in ipairs(obj) do
		sum = sum + val
	end

	return sum
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

	luaClient := app.script.c
	luaClient.db = db

	l := app.script.l

	err := app.script.l.DoString(testScript1)
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

	err = app.script.l.DoString(testScript2)
	if err != nil {
		t.Fatal(err)
	}

	v = luaReplyToLedisReply(l)
	if vv, ok := v.(string); !ok || vv != "PONG" {
		t.Fatal(fmt.Sprintf("%v %T", v, v))
	}

	err = app.script.l.DoString(testScript3)
	if err != nil {
		t.Fatal(err)
	}

	if v, err := db.Get([]byte("1")); err != nil {
		t.Fatal(err)
	} else if string(v) != "a" {
		t.Fatal(string(v))
	}

	err = app.script.l.DoString(testScript4)
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

	err = app.script.l.DoString(testScript5)
	if err != nil {
		t.Fatal(err)
	}

	v = luaReplyToLedisReply(l)
	if vv := v.(string); vv != "PONG" {
		t.Fatal(fmt.Sprintf("%v %T", v, v))
	}

	err = app.script.l.DoString(testScript6)
	if err != nil {
		t.Fatal(err)
	}

	v = luaReplyToLedisReply(l)
	vv := v.([]interface{})
	expected := []string{
		"string",
		"boolean",
		"string",
		"boolean",
		"nil",
	}
	if len(expected) != len(vv) {
		t.Fatalf("length different: %d, %d", len(expected), len(vv))
	}
	for i, r := range vv {
		s, ok := r.([]byte)
		if !ok {
			t.Errorf("reply[%d] expected: %s (%T), actual: %v (%T)",
				i, expected[i], expected[i], r, r)
		} else if string(s) != expected[i] {
			t.Errorf("reply[%d] expected: %s, actual: %v", i, expected[i], string(s))
		}
	}

	err = app.script.l.DoString(testScript7)
	if err != nil {
		t.Fatal(err)
	}

	v = luaReplyToLedisReply(l)
	if vv := v.(int64); vv != 42 {
		t.Fatal(fmt.Sprintf("%v %T", v, v))
	}

	err = app.script.l.DoString(testScript8)
	if err != nil {
		t.Fatal(err)
	}

	v = luaReplyToLedisReply(l)
	if vv := v.(int64); vv != 42 {
		t.Fatal(fmt.Sprintf("%v %T", v, v))
	}

	luaClient.db = nil
}
