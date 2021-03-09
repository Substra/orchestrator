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

type testCase struct {
	objective *NewObjective
	valid     bool
}

func TestObjectiveValidate(t *testing.T) {
	validAddressable := &Addressable{
		StorageAddress: "https://somewhere",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	validPerms := &NewPermissions{
		Public:        false,
		AuthorizedIds: []string{"org1"},
	}

	cases := map[string]testCase{
		"emtpy": {&NewObjective{}, false},
		"invalidKey": {&NewObjective{
			Key:            "not36chars",
			Name:           "invalid key",
			MetricsName:    "Test metric",
			Metrics:        validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, false},
		"valid": {&NewObjective{
			Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Name:           "Test objective",
			MetricsName:    "test metric",
			Metrics:        validAddressable,
			Description:    validAddressable,
			NewPermissions: validPerms,
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.objective.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.objective.Validate(), name+" should be invalid")
		}
	}
}
