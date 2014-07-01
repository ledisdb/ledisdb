package main

//#include <stdlib.h>
//#include "linenoise.h"
import "C"

import (
	"errors"
	"unsafe"
)

func line(prompt string) (string, error) {
	promptCString := C.CString(prompt)
	resultCString := C.linenoise(promptCString)
	C.free(unsafe.Pointer(promptCString))
	defer C.free(unsafe.Pointer(resultCString))

	if resultCString == nil {
		return "", errors.New("quited by a signal")
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
