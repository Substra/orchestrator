package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionsValue(t *testing.T) {
	permissions := &Permissions{
		Process:  &Permission{Public: true, AuthorizedIds: []string{"1", "2"}},
		Download: &Permission{Public: false, AuthorizedIds: []string{"4", "5"}},
	}

	value, err := permissions.Value()
	assert.NoError(t, err, "permissions serialization should not fail")

	scanned := new(Permissions)
	err = scanned.Scan(value)
	assert.NoError(t, err, "permissions scan should not fail")

	assert.Equal(t, permissions, scanned)
}

func TestPermissionValue(t *testing.T) {
	permission := &Permission{Public: true, AuthorizedIds: []string{"1", "2"}}

	value, err := permission.Value()
	assert.NoError(t, err, "permission serialization should not fail")

	scanned := new(Permission)
	err = scanned.Scan(value)
	assert.NoError(t, err, "permission scan should not fail")

	assert.Equal(t, permission, scanned)
}

func TestAlgoCategoryValue(t *testing.T) {
	cat := AlgoCategory_ALGO_SIMPLE
	category := &cat

	value, err := category.Value()
	assert.NoError(t, err, "algo category serialization should not fail")

	scanned := new(AlgoCategory)
	err = scanned.Scan(value)
	assert.NoError(t, err, "algo category scan should not fail")

	assert.Equal(t, category, scanned)
}

func TestComputeTaskStatusValue(t *testing.T) {
	s := ComputeTaskStatus_STATUS_DOING
	status := &s

	value, err := status.Value()
	assert.NoError(t, err, "task status serialization should not fail")

	scanned := new(ComputeTaskStatus)
	err = scanned.Scan(value)
	assert.NoError(t, err, "task status scan should not fail")

	assert.Equal(t, status, scanned)
}

func TestComputeTaskCategoryValue(t *testing.T) {
	cat := ComputeTaskCategory_TASK_TRAIN
	category := &cat

	value, err := category.Value()
	assert.NoError(t, err, "task category serialization should not fail")

	scanned := new(ComputeTaskCategory)
	err = scanned.Scan(value)
	assert.NoError(t, err, "task category scan should not fail")

	assert.Equal(t, category, scanned)
}

func TestErrorTypeValue(t *testing.T) {
	e := ErrorType_ERROR_TYPE_EXECUTION
	errorType := &e

	value, err := errorType.Value()
	assert.NoError(t, err, "error type serialization should not fail")

	scanned := new(ErrorType)
	err = scanned.Scan(value)
	assert.NoError(t, err, "error type scan should not fail")

	assert.Equal(t, errorType, scanned)
}

func TestModelCategoryValue(t *testing.T) {
	cat := ModelCategory_MODEL_SIMPLE
	category := &cat

	value, err := category.Value()
	assert.NoError(t, err, "model category serialization should not fail")

	scanned := new(ModelCategory)
	err = scanned.Scan(value)
	assert.NoError(t, err, "model category scan should not fail")

	assert.Equal(t, category, scanned)
}

func TestDataSampleValue(t *testing.T) {
	datasample := &DataSample{
		Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		DataManagerKeys: []string{"9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		Owner:           "testOwner",
		TestOnly:        false,
	}

	value, err := datasample.Value()
	assert.NoError(t, err, "datasample serialization should not fail")

	scanned := new(DataSample)
	err = scanned.Scan(value)
	assert.NoError(t, err, "datasample scan should not fail")

	assert.Equal(t, datasample, scanned)
}

func TestDataManagerValue(t *testing.T) {
	datamanager := &DataManager{
		Name:  "test",
		Owner: "testOwner",
	}

	value, err := datamanager.Value()
	assert.NoError(t, err, "datamanager serialization should not fail")

	scanned := new(DataManager)
	err = scanned.Scan(value)
	assert.NoError(t, err, "datamanager scan should not fail")

	assert.Equal(t, datamanager, scanned)
}
