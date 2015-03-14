// +build lua

package lua

//#include <lua.h>
//#include <lauxlib.h>
//#include <lualib.h>
//#include <stdlib.h>
//#include "golua.h"
import "C"
import "unsafe"

type LuaError struct {
	code       int
	message    string
	stackTrace []LuaStackEntry
}

func (err *LuaError) Error() string {
	return err.message
}

func (err *LuaError) Code() int {
	return err.code
}

func (err *LuaError) StackTrace() []LuaStackEntry {
	return err.stackTrace
}

// luaL_argcheck
func (L *State) ArgCheck(cond bool, narg int, extramsg string) {
	if cond {
		Cextramsg := C.CString(extramsg)
		defer C.free(unsafe.Pointer(Cextramsg))
		C.luaL_argerror(L.s, C.int(narg), Cextramsg)
	}
}

// luaL_argerror
func (L *State) ArgError(narg int, extramsg string) int {
	Cextramsg := C.CString(extramsg)
	defer C.free(unsafe.Pointer(Cextramsg))
	return int(C.luaL_argerror(L.s, C.int(narg), Cextramsg))
}

// luaL_callmeta
func (L *State) CallMeta(obj int, e string) int {
	Ce := C.CString(e)
	defer C.free(unsafe.Pointer(Ce))
	return int(C.luaL_callmeta(L.s, C.int(obj), Ce))
}

// luaL_checkany
func (L *State) CheckAny(narg int) {
	C.luaL_checkany(L.s, C.int(narg))
}

// luaL_checkinteger
func (L *State) CheckInteger(narg int) int {
	return int(C.luaL_checkinteger(L.s, C.int(narg)))
}

// luaL_checknumber
func (L *State) CheckNumber(narg int) float64 {
	return float64(C.luaL_checknumber(L.s, C.int(narg)))
}

// luaL_checkstring
func (L *State) CheckString(narg int) string {
	var length C.size_t
	return C.GoString(C.luaL_checklstring(L.s, C.int(narg), &length))
}

// luaL_checkoption
//
// BUG(everyone_involved): not implemented
func (L *State) CheckOption(narg int, def string, lst []string) int {
	//TODO: complication: lst conversion to const char* lst[] from string slice
	return 0
}

// luaL_checktype
func (L *State) CheckType(narg int, t LuaValType) {
	C.luaL_checktype(L.s, C.int(narg), C.int(t))
}

// luaL_checkudata
func (L *State) CheckUdata(narg int, tname string) unsafe.Pointer {
	Ctname := C.CString(tname)
	defer C.free(unsafe.Pointer(Ctname))
	return unsafe.Pointer(C.luaL_checkudata(L.s, C.int(narg), Ctname))
}

// Executes file, returns nil for no errors or the lua error string on failure
func (L *State) DoFile(filename string) error {
	if r := L.LoadFile(filename); r != 0 {
		return &LuaError{r, L.ToString(-1), L.StackTrace()}
	}
	return L.Call(0, LUA_MULTRET)
}

// Executes the string, returns nil for no errors or the lua error string on failure
func (L *State) DoString(str string) error {
	if r := L.LoadString(str); r != 0 {
		return &LuaError{r, L.ToString(-1), L.StackTrace()}
	}
	return L.Call(0, LUA_MULTRET)
}

// Like DoString but panics on error
func (L *State) MustDoString(str string) {
	if err := L.DoString(str); err != nil {
		panic(err)
	}
}

// luaL_getmetafield
func (L *State) GetMetaField(obj int, e string) bool {
	Ce := C.CString(e)
	defer C.free(unsafe.Pointer(Ce))
	return C.luaL_getmetafield(L.s, C.int(obj), Ce) != 0
}

// luaL_getmetatable
func (L *State) LGetMetaTable(tname string) {
	Ctname := C.CString(tname)
	defer C.free(unsafe.Pointer(Ctname))
	C.lua_getfield(L.s, LUA_REGISTRYINDEX, Ctname)
}

// luaL_gsub
func (L *State) GSub(s string, p string, r string) string {
	Cs := C.CString(s)
	Cp := C.CString(p)
	Cr := C.CString(r)
	defer func() {
		C.free(unsafe.Pointer(Cs))
		C.free(unsafe.Pointer(Cp))
		C.free(unsafe.Pointer(Cr))
	}()

	return C.GoString(C.luaL_gsub(L.s, Cs, Cp, Cr))
}

// luaL_loadfile
func (L *State) LoadFile(filename string) int {
	Cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(Cfilename))
	return int(C.luaL_loadfile(L.s, Cfilename))
}

// luaL_loadstring
func (L *State) LoadString(s string) int {
	Cs := C.CString(s)
	defer C.free(unsafe.Pointer(Cs))
	return int(C.luaL_loadstring(L.s, Cs))
}

// luaL_newmetatable
func (L *State) NewMetaTable(tname string) bool {
	Ctname := C.CString(tname)
	defer C.free(unsafe.Pointer(Ctname))
	return C.luaL_newmetatable(L.s, Ctname) != 0
}

// luaL_newstate
func NewState() *State {
	ls := (C.luaL_newstate())
	L := newState(ls)
	return L
}

// luaL_openlibs
func (L *State) OpenLibs() {
	C.luaL_openlibs(L.s)
	C.clua_hide_pcall(L.s)
}

// luaL_optinteger
func (L *State) OptInteger(narg int, d int) int {
	return int(C.luaL_optinteger(L.s, C.int(narg), C.lua_Integer(d)))
}

// luaL_optnumber
func (L *State) OptNumber(narg int, d float64) float64 {
	return float64(C.luaL_optnumber(L.s, C.int(narg), C.lua_Number(d)))
}

// luaL_optstring
func (L *State) OptString(narg int, d string) string {
	var length C.size_t
	Cd := C.CString(d)
	defer C.free(unsafe.Pointer(Cd))
	return C.GoString(C.luaL_optlstring(L.s, C.int(narg), Cd, &length))
}

// luaL_ref
func (L *State) Ref(t int) int {
	return int(C.luaL_ref(L.s, C.int(t)))
}

// luaL_typename
func (L *State) LTypename(index int) string {
	return C.GoString(C.lua_typename(L.s, C.lua_type(L.s, C.int(index))))
}

// luaL_unref
func (L *State) Unref(t int, ref int) {
	C.luaL_unref(L.s, C.int(t), C.int(ref))
}

// luaL_where
func (L *State) Where(lvl int) {
	C.luaL_where(L.s, C.int(lvl))
}
