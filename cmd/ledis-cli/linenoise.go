package main

//#include <stdlib.h>
//#include "linenoise.h"
//#include "linenoiseCompletionCallbackHook.h"
import "C"

import (
	"errors"
	"unsafe"
)

func init() {
	C.linenoiseSetupCompletionCallbackHook()
}

func line(prompt string) (string, error) {
	promptCString := C.CString(prompt)
	resultCString := C.linenoise(promptCString)
	C.free(unsafe.Pointer(promptCString))
	defer C.free(unsafe.Pointer(resultCString))

	if resultCString == nil {
		return "", errors.New("exiting due to signal")
	}

	result := C.GoString(resultCString)

	return result, nil
}

func addHistory(line string) error {
	lineCString := C.CString(line)
	res := C.linenoiseHistoryAdd(lineCString)
	C.free(unsafe.Pointer(lineCString))
	if res != 1 {
		return errors.New("Could not add line to history.")
	}
	return nil
}

func setHistoryCapacity(capacity int) error {
	res := C.linenoiseHistorySetMaxLen(C.int(capacity))
	if res != 1 {
		return errors.New("Could not set history max len.")
	}
	return nil
}

// CompletionHandler provides possible completions for given input
type CompletionHandler func(input string) []string

// DefaultCompletionHandler simply returns an empty slice.
var DefaultCompletionHandler = func(input string) []string {
	return make([]string, 0)
}

var complHandler = DefaultCompletionHandler

// SetCompletionHandler sets the CompletionHandler to be used for completion
func SetCompletionHandler(c CompletionHandler) {
	complHandler = c
}

// typedef struct linenoiseCompletions {
//   size_t len;
//   char **cvec;
// } linenoiseCompletions;
// typedef void(linenoiseCompletionCallback)(const char *, linenoiseCompletions *);
// void linenoiseSetCompletionCallback(linenoiseCompletionCallback *);
// void linenoiseAddCompletion(linenoiseCompletions *, char *);

//export linenoiseGoCompletionCallbackHook
func linenoiseGoCompletionCallbackHook(input *C.char, completions *C.linenoiseCompletions) {
	completionsSlice := complHandler(C.GoString(input))

	completionsLen := len(completionsSlice)
	completions.len = C.size_t(completionsLen)

	if completionsLen > 0 {
		cvec := C.malloc(C.size_t(int(unsafe.Sizeof(*(**C.char)(nil))) * completionsLen))
		cvecSlice := (*(*[999999]*C.char)(cvec))[:completionsLen]

		for i, str := range completionsSlice {
			cvecSlice[i] = C.CString(str)
		}
		completions.cvec = (**C.char)(cvec)
	}
}
