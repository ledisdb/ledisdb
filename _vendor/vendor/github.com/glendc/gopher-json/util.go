package json

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/yuin/gopher-lua"
)

var (
	errFunction = errors.New("cannot encode function to JSON")
	errChannel  = errors.New("cannot encode channel to JSON")
	errState    = errors.New("cannot encode state to JSON")
	errUserData = errors.New("cannot encode userdata to JSON")
	errNested   = errors.New("cannot encode recursively nested tables to JSON")
)

type jsonValue struct {
	lua.LValue
	visited map[*lua.LTable]bool
}

func (j jsonValue) MarshalJSON() ([]byte, error) {
	return toJSON(j.LValue, j.visited)
}

func toJSON(value lua.LValue, visited map[*lua.LTable]bool) (data []byte, err error) {
	switch converted := value.(type) {
	case lua.LBool:
		data, err = json.Marshal(converted)
	case lua.LChannel:
		err = errChannel
	case lua.LNumber:
		data, err = json.Marshal(converted)
	case *lua.LFunction:
		err = errFunction
	case *lua.LNilType:
		data, err = json.Marshal(converted)
	case *lua.LState:
		err = errState
	case lua.LString:
		data, err = json.Marshal(converted)
	case *lua.LTable:
		var arr []jsonValue
		var obj map[string]jsonValue

		if visited[converted] {
			panic(errNested)
		}
		visited[converted] = true

		converted.ForEach(func(k lua.LValue, v lua.LValue) {
			i, numberKey := k.(lua.LNumber)
			if numberKey && obj == nil {
				index := int(i) - 1
				if index != len(arr) {
					// map out of order; convert to map
					obj = make(map[string]jsonValue)
					for i, value := range arr {
						obj[strconv.Itoa(i+1)] = value
					}
					obj[strconv.Itoa(index+1)] = jsonValue{v, visited}
					return
				}
				arr = append(arr, jsonValue{v, visited})
				return
			}
			if obj == nil {
				obj = make(map[string]jsonValue)
				for i, value := range arr {
					obj[strconv.Itoa(i+1)] = value
				}
			}
			obj[k.String()] = jsonValue{v, visited}
		})
		if obj != nil {
			data, err = json.Marshal(obj)
		} else {
			data, err = json.Marshal(arr)
		}
	case *lua.LUserData:
		// TODO: call metatable __tostring?
		err = errUserData
	}
	return
}

func fromJSON(L *lua.LState, value interface{}) lua.LValue {
	switch converted := value.(type) {
	case bool:
		return lua.LBool(converted)
	case float64:
		return lua.LNumber(converted)
	case string:
		return lua.LString(converted)
	case []interface{}:
		arr := L.CreateTable(len(converted), 0)
		for _, item := range converted {
			arr.Append(fromJSON(L, item))
		}
		return arr
	case map[string]interface{}:
		tbl := L.CreateTable(0, len(converted))
		for key, item := range converted {
			tbl.RawSetH(lua.LString(key), fromJSON(L, item))
		}
		return tbl
	}
	return lua.LNil
}
