// +build lua

package lua

/*
#include <lua.h>
#include <lualib.h>
#include <stdlib.h>
*/
import "C"

import (
	"reflect"
	"unsafe"
)

// Type of allocation functions to use with NewStateAlloc
type Alloc func(ptr unsafe.Pointer, osize uint, nsize uint) unsafe.Pointer

// This is the type of go function that can be registered as lua functions
type LuaGoFunction func(L *State) int

// Wrapper to keep cgo from complaining about incomplete ptr type
//export State
type State struct {
	// Wrapped lua_State object
	s *C.lua_State

	// Registry of go object that have been pushed to Lua VM
	registry []interface{}

	// Freelist for funcs indices, to allow for freeing
	freeIndices []uint
}

//export golua_callgofunction
func golua_callgofunction(L interface{}, fid uint) int {
	L1 := L.(*State)
	if fid < 0 {
		panic(&LuaError{0, "Requested execution of an unknown function", L1.StackTrace()})
	}
	f := L1.registry[fid].(LuaGoFunction)
	return f(L1)
}

//export golua_interface_newindex_callback
func golua_interface_newindex_callback(Li interface{}, iid uint, field_name_cstr *C.char) int {
	L := Li.(*State)
	iface := L.registry[iid]
	ifacevalue := reflect.ValueOf(iface).Elem()

	field_name := C.GoString(field_name_cstr)

	fval := ifacevalue.FieldByName(field_name)

	if fval.Kind() == reflect.Ptr {
		fval = fval.Elem()
	}

	luatype := LuaValType(C.lua_type(L.s, 3))

	switch fval.Kind() {
	case reflect.Bool:
		if luatype == LUA_TBOOLEAN {
			fval.SetBool(int(C.lua_toboolean(L.s, 3)) != 0)
			return 1
		} else {
			L.PushString("Wrong assignment to field " + field_name)
			return -1
		}

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		if luatype == LUA_TNUMBER {
			fval.SetInt(int64(C.lua_tointeger(L.s, 3)))
			return 1
		} else {
			L.PushString("Wrong assignment to field " + field_name)
			return -1
		}

	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		if luatype == LUA_TNUMBER {
			fval.SetUint(uint64(C.lua_tointeger(L.s, 3)))
			return 1
		} else {
			L.PushString("Wrong assignment to field " + field_name)
			return -1
		}

	case reflect.String:
		if luatype == LUA_TSTRING {
			fval.SetString(C.GoString(C.lua_tolstring(L.s, 3, nil)))
			return 1
		} else {
			L.PushString("Wrong assignment to field " + field_name)
			return -1
		}

	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		if luatype == LUA_TNUMBER {
			fval.SetFloat(float64(C.lua_tonumber(L.s, 3)))
			return 1
		} else {
			L.PushString("Wrong assignment to field " + field_name)
			return -1
		}
	}

	L.PushString("Unsupported type of field " + field_name + ": " + fval.Type().String())
	return -1
}

//export golua_interface_index_callback
func golua_interface_index_callback(Li interface{}, iid uint, field_name *C.char) int {
	L := Li.(*State)
	iface := L.registry[iid]
	ifacevalue := reflect.ValueOf(iface).Elem()

	fval := ifacevalue.FieldByName(C.GoString(field_name))

	if fval.Kind() == reflect.Ptr {
		fval = fval.Elem()
	}

	switch fval.Kind() {
	case reflect.Bool:
		L.PushBoolean(fval.Bool())
		return 1

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		L.PushInteger(fval.Int())
		return 1

	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		L.PushInteger(int64(fval.Uint()))
		return 1

	case reflect.String:
		L.PushString(fval.String())
		return 1

	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		L.PushNumber(fval.Float())
		return 1
	}

	L.PushString("Unsupported type of field: " + fval.Type().String())
	return -1
}

//export golua_gchook
func golua_gchook(L interface{}, id uint) int {
	L1 := L.(*State)
	L1.unregister(id)
	return 0
}

//export golua_callpanicfunction
func golua_callpanicfunction(L interface{}, id uint) int {
	L1 := L.(*State)
	f := L1.registry[id].(LuaGoFunction)
	return f(L1)
}

//export golua_idtointerface
func golua_idtointerface(id uint) interface{} {
	return id
}

//export golua_cfunctiontointerface
func golua_cfunctiontointerface(f *uintptr) interface{} {
	return f
}

//export golua_callallocf
func golua_callallocf(fp uintptr, ptr uintptr, osize uint, nsize uint) uintptr {
	return uintptr((*((*Alloc)(unsafe.Pointer(fp))))(unsafe.Pointer(ptr), osize, nsize))
}

//export go_panic_msghandler
func go_panic_msghandler(Li interface{}, z *C.char) {
	L := Li.(*State)
	s := C.GoString(z)

	panic(&LuaError{LUA_ERRERR, s, L.StackTrace()})
}
