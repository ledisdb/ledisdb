package json

import "github.com/yuin/gopher-lua"

const (
	// CJsonLibName defines the name for the cjson lib
	CJsonLibName = "cjson"
	// https://www.kyne.com.au/~mark/software/lua-cjson-manual.html
	cjsonLibVersion = "0.1.0"
)

// OpenCJSON opens the cjson lib
func OpenCJSON(l *lua.LState) int {
	mod := l.RegisterModule(CJsonLibName, cjsonFuncs).(*lua.LTable)
	mod.RawSetString("_NAME", lua.LString(CJsonLibName))
	mod.RawSetString("_VERSION", lua.LString(cjsonLibVersion))
	mod.RawSetString("null", lua.LNil)
	l.Push(mod)
	return 1
}

var cjsonFuncs = map[string]lua.LGFunction{
	"decode":                  apiDecode,
	"decode_invalid_numbers":  apiDecodeInvalidNumbers,
	"decode_max_depth":        apiDecodeMaxDepth,
	"encode":                  apiEncode,
	"encode_invalid_numbers":  apiEncodeInvalidNumbers,
	"encode_keep_buffer":      apiEncodeKeepBuffer,
	"encode_max_depth":        apiEncodeMaxDepth,
	"encode_number_precision": apiEncodeNumberPrecision,
	"encode_sparse_array":     apiEncodeSparseArray,
}

// settings functions

func apiDecodeInvalidNumbers(l *lua.LState) int {
	l.RaiseError("decode_invalid_numbers is not supported")
	return 0
}

func apiDecodeMaxDepth(l *lua.LState) int {
	l.RaiseError("decode_max_depth is not supported")
	return 0
}

func apiEncodeInvalidNumbers(l *lua.LState) int {
	l.RaiseError("encode_invalid_numbers is not supported")
	return 0
}

func apiEncodeKeepBuffer(l *lua.LState) int {
	l.RaiseError("encode_keep_buffer is not supported")
	return 0
}

func apiEncodeMaxDepth(l *lua.LState) int {
	l.RaiseError("encode_max_depth is not supported")
	return 0
}

func apiEncodeNumberPrecision(l *lua.LState) int {
	l.RaiseError("encode_number_precision is not supported")
	return 0
}

func apiEncodeSparseArray(l *lua.LState) int {
	l.RaiseError("encode_sparse_array is not supported")
	return 0
}
