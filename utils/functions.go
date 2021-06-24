package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// GetCaller returns the name of the function calling it.
// The argument skip is the number of stack frames
// to ascend, with 0 identifying the caller of GetCaller
func GetCaller(skip int) (string, error) {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", fmt.Errorf("failed to call runtime.Caller")
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "", fmt.Errorf("runtime.FuncForPC returned nil")
	}

	fnName := fn.Name()
	idx := strings.LastIndex(fnName, ".")
	if idx == -1 {
		return "", fmt.Errorf("cannot find . in function name %s", fnName)
	}

	return fnName[idx+1:], nil
}
