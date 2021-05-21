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

package concurrency

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var excludedSuffixes []string = []string{
	"_test.go",
}

// This static check ensures that all the methods in server/standalone/*.go
// implement the scheduler token wait mechanism:
//
//   execToken := <-s.scheduler.AcquireExecutionToken()
//   defer execToken.Release()
func TestServerMethodsUseScheduler(t *testing.T) {
	files := getAllFiles(t)
	for _, f := range files {
		validateSourceFile(t, f.Name())
	}
}

func getAllFiles(t *testing.T) []os.FileInfo {
	res := []os.FileInfo{}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		assert.True(t, ok, "Should successfully get caller information")
	}

	dir := path.Dir(filename)
	files, err := ioutil.ReadDir(dir)
	assert.NoError(t, err)

	for _, f := range files {
		if isEligibleFile(f) {
			res = append(res, f)
		}
	}

	return res
}

func isEligibleFile(f os.FileInfo) bool {
	// Must not be a directory
	if f.IsDir() {
		return false
	}

	// Must not finish with _test.go or _interceptor.go
	for _, suffix := range excludedSuffixes {
		if strings.HasSuffix(f.Name(), suffix) {
			return false
		}
	}

	// Eligible!
	return true
}

func validateSourceFile(t *testing.T, path string) {
	fset := token.NewFileSet()

	dat, err := ioutil.ReadFile(path)
	assert.NoError(t, err)

	f, err := parser.ParseFile(fset, "", string(dat), 0)
	assert.NoError(t, err)

	ast.Inspect(f, func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			funcName := fn.Name.Name
			if isSmartContractMethod(fn) {
				ok := validateMethod(fn)
				assert.True(t, ok, fmt.Sprintf("method %s should wait for the execution token", funcName))
				return ok
			}
		}
		return true
	})
}

func isSmartContractMethod(fn *ast.FuncDecl) bool {
	if len(fn.Type.Params.List) == 0 {
		// Function must have at least one parameter
		return false
	}

	param, ok := fn.Type.Params.List[0].Type.(*ast.SelectorExpr)
	if !ok {
		// The first param node must be of type SelectorExpr
		return false
	}

	if param.Sel.Name != "Context" {
		// The first param must be of type "Context"
		return false
	}

	// Looks like a smart contract method
	return true
}

func validateMethod(f *ast.FuncDecl) bool {

	if len(f.Body.List) == 0 {
		// Method body must not be empty
		return false
	}

	as, ok := f.Body.List[0].(*ast.AssignStmt)
	if !ok {
		// The first statement must be an assignment
		return false
	}

	lhs, ok := as.Lhs[0].(*ast.Ident)
	if !ok {
		// The LHS in the first statement must be an Ident
		return false
	}

	if lhs.Name != "execToken" {
		// The LHS in the first statement must be named execToken
		return false
	}

	// Valid!
	return true
}
