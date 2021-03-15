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
		"conflict":   {err: errors.ErrConflict, code: codes.AlreadyExists},
		"validation": {err: errors.ErrInvalidAsset, code: codes.InvalidArgument},
		"unknown":    {err: fmt.Errorf("some unknown error"), code: codes.Unknown},
	}

	for name, tc := range cases {
		status := status.Convert(toStatus(tc.err))
		assert.Equal(t, tc.code, status.Code(), fmt.Sprintf("Code conversion should match for %s", name))
	}

	assert.Nil(t, toStatus(nil), "nil should not be mapped")
}
