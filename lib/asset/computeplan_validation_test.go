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

package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNewComputePlan(t *testing.T) {
	cases := map[string]struct {
		newComputePlan *NewComputePlan
		valid          bool
	}{
		"emtpy": {&NewComputePlan{}, false},
		"invalidKey": {&NewComputePlan{
			Key: "not36chars",
		}, false},
		"valid": {&NewComputePlan{
			Key: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.newComputePlan.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.newComputePlan.Validate(), name+" should be invalid")
		}
	}
}
