package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterDataManager(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	mps := new(MockPermissionAPI)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)

	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	service := NewDataManagerService(provider)

	newPerms := &asset.NewPermissions{Public: true}
	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	newLogsPerm := &asset.NewPermissions{Public: true}
	logsPerm := &asset.Permission{Public: true}

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	opener := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	newDataManager := &asset.NewDataManager{
		Key:            "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:           "test datamanager",
		Description:    description,
		Opener:         opener,
		Type:           "test dm",
		NewPermissions: newPerms,
		LogsPermission: newLogsPerm,
	}

	storedDataManager := &asset.DataManager{
		Key:            "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:           "test datamanager",
		Owner:          "owner",
		Permissions:    perms,
		Description:    description,
		Opener:         opener,
		Type:           "test dm",
		CreationDate:   timestamppb.New(time.Unix(1337, 0)),
		LogsPermission: logsPerm,
	}

	mps.On("CreatePermission", "owner", newPerms).Return(&asset.Permission{Public: true}, nil).Once()
	mps.On("CreatePermissions", "owner", newPerms).Return(perms, nil).Once()
	dbal.On("DataManagerExists", newDataManager.GetKey()).Return(false, nil).Once()
	dbal.On("AddDataManager", storedDataManager).Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
		AssetKey:  storedDataManager.Key,
		Asset:     &asset.Event_DataManager{DataManager: storedDataManager},
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	dm, err := service.RegisterDataManager(newDataManager, "owner")

	assert.NoError(t, err, "Registration of valid datamanager should not fail")
	assert.NotNil(t, dm, "Registration should return a datamanager asset")
	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestGetDataManager(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetDataManagerDBAL").Return(dbal)
	service := NewDataManagerService(provider)

	datamanager := asset.DataManager{
		Key:  "objKey",
		Name: "Test",
	}

	dbal.On("GetDataManager", "objKey").Return(&datamanager, nil).Once()

	o, err := service.GetDataManager("objKey")
	require.Nil(t, err)
	assert.Equal(t, o.Name, datamanager.Name)
	dbal.AssertExpectations(t)
}

func TestQueryDataManagers(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetDataManagerDBAL").Return(dbal)
	service := NewDataManagerService(provider)

	dm1 := asset.DataManager{
		Key:  "obj1",
		Name: "Test 1",
	}
	dm2 := asset.DataManager{
		Key:  "obj2",
		Name: "Test 2",
	}

	pagination := common.NewPagination("", 12)

	dbal.On("QueryDataManagers", pagination).Return([]*asset.DataManager{&dm1, &dm2}, "nextPage", nil).Once()

	r, token, err := service.QueryDataManagers(pagination)
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, dm1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
	dbal.AssertExpectations(t)
}

func TestIsOwner(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	provider.On("GetDataManagerDBAL").Return(dbal)
	service := NewDataManagerService(provider)

	dm := &asset.DataManager{
		Key:   "obj1",
		Name:  "Test 1",
		Owner: "owner",
	}

	dbal.On("GetDataManager", "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6").Return(dm, nil).Once()
	ok, err := service.isOwner("65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6", "owner")

	assert.NoError(t, err, "is owner should not fail")
	assert.Equal(t, ok, true, "owner owns the datamanager")
	dbal.AssertExpectations(t)
}

func TestCheckDataManager(t *testing.T) {
	type result struct {
		dataManager *asset.DataManager
		err         string
	}

	dataManager := &asset.DataManager{Key: "uuid1"}
	dataSampleKeys := []string{"uuid2", "uuid3"}
	owner := "owner"

	cases := []struct {
		name                string
		canProcess          bool
		checkSameManagerErr error
		result              result
	}{
		{
			name:       "ok",
			canProcess: true,
			result:     result{dataManager: dataManager},
		},
		{
			name:       "not authorized",
			canProcess: false,
			result:     result{err: "not authorized to process datamanager"},
		},
		{
			name:                "not same data manager",
			canProcess:          true,
			checkSameManagerErr: errors.New("datasamples do not share a common manager"),
			result:              result{err: "datasamples do not share a common manager"},
		},
	}

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {
				dbal := new(persistence.MockDBAL)
				ps := new(MockPermissionAPI)
				dss := new(MockDataSampleAPI)
				provider := newMockedProvider()

				ps.On("CanProcess", dataManager.Permissions, owner).Return(c.canProcess)
				dss.On("CheckSameManager", dataManager.Key, dataSampleKeys).Return(c.checkSameManagerErr)

				provider.On("GetDataManagerDBAL").Return(dbal)
				provider.On("GetPermissionService").Return(ps)
				provider.On("GetDataSampleService").Return(dss)

				service := NewDataManagerService(provider)

				err := service.CheckDataManager(dataManager, dataSampleKeys, owner)

				if c.result.err == "" {
					assert.NoError(t, err)
				} else {
					assert.ErrorContains(t, err, c.result.err)
				}
			},
		)
	}
}

func TestUpdateSingleExistingDataManager(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	provider.On("GetEventService").Return(es)
	provider.On("GetDataManagerDBAL").Return(dbal)
	service := NewDataManagerService(provider)

	existingDataManager := &asset.DataManager{
		Key:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name:  "data manager name",
		Owner: "owner",
	}

	updateDataManagerParam := &asset.UpdateDataManagerParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated data manager name",
	}

	storedDataManager := &asset.DataManager{
		Key:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name:  "Updated data manager name",
		Owner: "owner",
	}

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
		AssetKey:  storedDataManager.Key,
		Asset:     &asset.Event_DataManager{DataManager: storedDataManager},
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
			dbal.On("GetDataManager", existingDataManager.GetKey()).Return(existingDataManager, nil).Once()

			if tc.valid {
				dbal.On("UpdateDataManager", storedDataManager).Return(nil).Once()
				es.On("RegisterEvents", e).Once().Return(nil)
			}

			err := service.UpdateDataManager(updateDataManagerParam, tc.requester)

			if tc.valid {
				assert.NoError(t, err, "Update of data manager should not fail")
			} else {
				assert.Error(t, err, "Update of data manager should fail")
			}

			dbal.AssertExpectations(t)
			es.AssertExpectations(t)
		})
	}
}
