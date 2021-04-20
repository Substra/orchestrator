// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
