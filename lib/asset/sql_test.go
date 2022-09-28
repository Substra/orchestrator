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

func TestAssetKindValue(t *testing.T) {
	k := AssetKind_ASSET_ORGANIZATION
	kind := &k

	value, err := kind.Value()
	assert.NoError(t, err, "asset kind serialization should not fail")

	scanned := new(AssetKind)
	err = scanned.Scan(value)
	assert.NoError(t, err, "asset kind scan should not fail")

	assert.Equal(t, kind, scanned)
}

func TestEventKindValue(t *testing.T) {
	k := EventKind_EVENT_ASSET_CREATED
	kind := &k

	value, err := kind.Value()
	assert.NoError(t, err, "event kind serialization should not fail")

	scanned := new(EventKind)
	err = scanned.Scan(value)
	assert.NoError(t, err, "event kind scan should not fail")

	assert.Equal(t, kind, scanned)
}
