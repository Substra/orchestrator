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

package common

import (
	"fmt"
	"testing"

	"github.com/owkin/orchestrator/lib/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStatusConversion(t *testing.T) {
	cases := map[string]struct {
		err  error
		code codes.Code
	}{
		"conflict":            {err: errors.ErrConflict, code: codes.AlreadyExists},
		"validation":          {err: errors.ErrInvalidAsset, code: codes.InvalidArgument},
		"unknown":             {err: fmt.Errorf("some unknown error"), code: codes.Unknown},
		"unauthorized":        {err: errors.ErrPermissionDenied, code: codes.PermissionDenied},
		"invalid_reference":   {err: errors.ErrReferenceNotFound, code: codes.InvalidArgument},
		"notfound":            {err: errors.ErrNotFound, code: codes.NotFound},
		"badrequest":          {err: errors.ErrBadRequest, code: codes.FailedPrecondition},
		"incompatible_status": {err: errors.ErrIncompatibleTaskStatus, code: codes.InvalidArgument},
		"unimplemented":       {err: errors.ErrUnimplemented, code: codes.Unimplemented},
		"unprocessable model": {err: errors.ErrCannotDisableModel, code: codes.InvalidArgument},
	}

	for name, tc := range cases {
		t.Run(fmt.Sprintf("fromError(%s)", name), func(t *testing.T) {
			assert.Equal(t, tc.code, status.Convert(fromError(tc.err)).Code())
		})
		err := fmt.Errorf("new error with embedded code: %s in the message", tc.err.Error())
		t.Run(fmt.Sprintf("fromMessage(%s)", name), func(t *testing.T) {
			assert.Equal(t, tc.code, status.Convert(fromMessage(err.Error())).Code())
		})
	}

	assert.Nil(t, fromError(nil), "nil should not be mapped")
}
