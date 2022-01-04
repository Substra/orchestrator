package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type failureReportTestCase struct {
	failureReport *NewFailureReport
	valid         bool
}

func TestFailureReportValidate(t *testing.T) {
	validAddressable := &Addressable{
		StorageAddress: "https://somewhere",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	cases := map[string]failureReportTestCase{
		"empty": {&NewFailureReport{}, false},
		"invalidComputeTaskKey": {&NewFailureReport{
			ComputeTaskKey: "notUUID",
			LogsAddress:    validAddressable,
		}, false},
		"valid": {&NewFailureReport{
			ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
			LogsAddress:    validAddressable,
		}, true},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.failureReport.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.failureReport.Validate(), name+" should be invalid")
		}
	}
}
