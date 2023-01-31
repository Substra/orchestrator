package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterFunction(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	mps := new(MockPermissionAPI)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()

	provider.On("GetFunctionDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)

	service := NewFunctionService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	functionLocation := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	newPerms := &asset.NewPermissions{Public: true}

	function := &asset.NewFunction{
		Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:           "Test function",
		Function:      functionLocation,
		Description:    description,
		NewPermissions: newPerms,
	}

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}
	mps.On("CreatePermissions", "owner", newPerms).Return(perms, nil).Once()

	dbal.On("FunctionExists", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(false, nil).Once()

	storedFunction := &asset.Function{
		Key:          "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:         "Test function",
		Function:    functionLocation,
		Description:  description,
		Permissions:  perms,
		Owner:        "owner",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}
	dbal.On("AddFunction", storedFunction).Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		AssetKey:  function.Key,
		Asset:     &asset.Event_Function{Function: storedFunction},
	}
	es.On("RegisterEvents", e).Return(nil)

	o, err := service.RegisterFunction(function, "owner")

	assert.NoError(t, err, "Registration of valid function should not fail")
	assert.NotNil(t, o, "Registration should return an Function")
	assert.Equal(t, perms, o.Permissions, "Permissions should be set")
	assert.Equal(t, "owner", o.Owner, "Owner should be set")

	dbal.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestGetFunction(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetFunctionDBAL").Return(dbal)
	service := NewFunctionService(provider)

	function := asset.Function{
		Key:  "functionKey",
		Name: "Test",
	}

	dbal.On("GetFunction", "functionKey").Return(&function, nil).Once()

	o, err := service.GetFunction("functionKey")
	require.Nil(t, err)
	assert.Equal(t, o.Name, function.Name)
}

func TestQueryFunctions(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetFunctionDBAL").Return(dbal)
	service := NewFunctionService(provider)

	computePlanKey := uuid.NewString()

	function1 := asset.Function{
		Key:  "function1",
		Name: "Test 1",
	}
	function2 := asset.Function{
		Key:  "function2",
		Name: "Test 2",
	}

	pagination := common.NewPagination("", 12)

	dbal.On("QueryFunctions", pagination, &asset.FunctionQueryFilter{ComputePlanKey: computePlanKey}).
		Return([]*asset.Function{&function1, &function2}, "nextPage", nil).Once()

	r, token, err := service.QueryFunctions(pagination, &asset.FunctionQueryFilter{ComputePlanKey: computePlanKey})
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, function1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}

func TestCanDownload(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetFunctionDBAL").Return(dbal)
	service := NewFunctionService(provider)

	perms := &asset.Permissions{
		Process: &asset.Permission{Public: true},
		Download: &asset.Permission{
			Public:        false,
			AuthorizedIds: []string{"org-2"},
		},
	}

	function := &asset.Function{
		Key:         "837B2E87-35CA-48F9-B83C-B40FB3FBA4E6",
		Name:        "Test",
		Permissions: perms,
	}

	dbal.On("GetFunction", "obj1").Return(function, nil).Once()

	ok, err := service.CanDownload("obj1", "org-2")

	assert.Equal(t, ok, true)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
}

func TestUpdateSingleExistingFunction(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	provider.On("GetEventService").Return(es)
	provider.On("GetFunctionDBAL").Return(dbal)
	service := NewFunctionService(provider)

	existingFunction := &asset.Function{
		Key:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name:  "function name",
		Owner: "owner",
	}

	updateFunctionParam := &asset.UpdateFunctionParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated function name",
	}

	storedFunction := &asset.Function{
		Key:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name:  "Updated function name",
		Owner: "owner",
	}

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		AssetKey:  storedFunction.Key,
		Asset:     &asset.Event_Function{Function: storedFunction},
	}

	cases := map[string]struct {
		requester string
		valid     bool
	}{
		"update successful": {
			requester: "owner",
			valid:     true,
		},
		"update rejected: requester is not owner": {
			requester: "user",
			valid:     false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dbal.On("GetFunction", existingFunction.GetKey()).Return(existingFunction, nil).Once()

			if tc.valid {
				dbal.On("UpdateFunction", storedFunction).Return(nil).Once()
				es.On("RegisterEvents", e).Once().Return(nil)
			}

			err := service.UpdateFunction(updateFunctionParam, tc.requester)

			if tc.valid {
				assert.NoError(t, err, "Update of function should not fail")
			} else {
				assert.Error(t, err, "Update of function should fail")
			}

			dbal.AssertExpectations(t)
			es.AssertExpectations(t)
		})
	}

}
