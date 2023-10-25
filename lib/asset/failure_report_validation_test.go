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
		"invalidAssetKeyFunction": {&NewFailureReport{
			AssetKey:    "notUUID",
			AssetType:   FailedAssetKind_FAILED_ASSET_FUNCTION,
			ErrorType:   ErrorType_ERROR_TYPE_BUILD,
			LogsAddress: nil,
		}, false},
		"invalidAssetKeyComputeTask": {&NewFailureReport{
			AssetKey:    "notUUID",
			AssetType:   FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
			ErrorType:   ErrorType_ERROR_TYPE_BUILD,
			LogsAddress: nil,
		}, false},
		"validBuildError": {&NewFailureReport{
			AssetKey:    "08680966-97ae-4573-8b2d-6c4db2b3c532",
			AssetType:   FailedAssetKind_FAILED_ASSET_FUNCTION,
			ErrorType:   ErrorType_ERROR_TYPE_BUILD,
			LogsAddress: validAddressable,
		}, true},
		"invalidBuildError": {&NewFailureReport{
			AssetKey:    "08680966-97ae-4573-8b2d-6c4db2b3c532",
			AssetType:   FailedAssetKind_FAILED_ASSET_FUNCTION,
			ErrorType:   ErrorType_ERROR_TYPE_BUILD,
			LogsAddress: nil,
		}, false},
		"validExecutionError": {&NewFailureReport{
			AssetKey:    "08680966-97ae-4573-8b2d-6c4db2b3c532",
			AssetType:   FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
			ErrorType:   ErrorType_ERROR_TYPE_EXECUTION,
			LogsAddress: validAddressable,
		}, true},
		"invalidExecutionError": {&NewFailureReport{
			AssetKey:    "08680966-97ae-4573-8b2d-6c4db2b3c532",
			AssetType:   FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
			ErrorType:   ErrorType_ERROR_TYPE_EXECUTION,
			LogsAddress: nil,
		}, false},
		"validInternalError": {&NewFailureReport{
			AssetKey:    "08680966-97ae-4573-8b2d-6c4db2b3c532",
			AssetType:   FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
			ErrorType:   ErrorType_ERROR_TYPE_INTERNAL,
			LogsAddress: nil,
		}, true},
		"invalidInternalError": {&NewFailureReport{
			AssetKey:    "08680966-97ae-4573-8b2d-6c4db2b3c532",
			AssetType:   FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
			ErrorType:   ErrorType_ERROR_TYPE_INTERNAL,
			LogsAddress: validAddressable,
		}, false},
	}

	for name, tc := range cases {
		if tc.valid {
			assert.NoError(t, tc.failureReport.Validate(), name+" should be valid")
		} else {
			assert.Error(t, tc.failureReport.Validate(), name+" should be invalid")
		}
	}
}
