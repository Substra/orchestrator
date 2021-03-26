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

type datasampleTestCase struct {
	datasample *NewDataSample
	valid      bool
}

type updateDataSampleTestCase struct {
	datasample *DataSampleUpdateParam
	valid      bool
}

func TestNewDataSampleValidate(t *testing.T) {
	cases := map[string]datasampleTestCase{
		"empty": {&NewDataSample{}, false},
		"invalidDataSampleKey": {&NewDataSample{
			Keys:            []string{"not36chars", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
			DataManagerKeys: []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "c6cc913d-83a9-4a8e-a258-2901e1d5ebbc"},
			TestOnly:        false,
		}, false},
		"invalidDataManagerKey": {&NewDataSample{
			Keys:            []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "c6cc913d-83a9-4a8e-a258-2901e1d5ebbc"},
			DataManagerKeys: []string{"not36chars", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
			TestOnly:        false,
		}, false},
		"valid": {&NewDataSample{
			Keys:            []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "c6cc913d-83a9-4a8e-a258-2901e1d5ebbc"},
			DataManagerKeys: []string{"3dd165f8-8822-481a-8bf9-23bf135152cf", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
			TestOnly:        false,
		}, true},
		"validTestOnly": {&NewDataSample{
			Keys:            []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "c6cc913d-83a9-4a8e-a258-2901e1d5ebbc"},
			DataManagerKeys: []string{"3dd165f8-8822-481a-8bf9-23bf135152cf", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
			TestOnly:        true,
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.datasample.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.datasample.Validate(), name+" should be invalid")
		}
	}
}

func TestUpdateDataSampleValidate(t *testing.T) {
	cases := map[string]updateDataSampleTestCase{
		"empty": {&DataSampleUpdateParam{}, false},
		"invalidDataSampleKey": {&DataSampleUpdateParam{
			Keys:            []string{"not36chars", "08680966-97ae-4573-8b2d-6c4db2b3cdd2"},
			DataManagerKeys: []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "08680966-97ae-4573-8b2d-6c4db2b3cdd2"},
		}, false},
		"invalidDataManagerKey": {&DataSampleUpdateParam{
			Keys:            []string{"08680966-97ae-4573-8b2d-6c4db2b3cdd2", "3dd165f8-8822-481a-8bf9-23bf135152cf"},
			DataManagerKeys: []string{"not36chars", "08680966-97ae-4573-8b2d-6c4db2b3cdd2"},
		}, false},
		"valid": {&DataSampleUpdateParam{
			Keys:            []string{"834f47c3-2d95-4ccd-a718-7143b64e61c0", "c6cc913d-83a9-4a8e-a258-2901e1d5ebbc"},
			DataManagerKeys: []string{"3dd165f8-8822-481a-8bf9-23bf135152cf", "1d417d76-a2e1-46e7-aae5-9c7c165575fc"},
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.datasample.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.datasample.Validate(), name+" should be invalid")
		}
	}
}
