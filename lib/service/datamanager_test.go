package service

import (
	"errors"
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGetCheckedDataManager(t *testing.T) {
	type getDataManagerResponse struct {
		dataManager *asset.DataManager
		err         error
	}

	type result struct {
		dataManager *asset.DataManager
		err         string
	}

	cases := []struct {
		name                string
		dataManagerResponse getDataManagerResponse
		canProcess          bool
		checkSameManagerErr error
		result              result
	}{
		{
			name:                "ok",
			dataManagerResponse: getDataManagerResponse{dataManager: &asset.DataManager{}},
			canProcess:          true,
			result:              result{dataManager: &asset.DataManager{}},
		},
		{
			name:                "data manager doesn't exist",
			dataManagerResponse: getDataManagerResponse{err: errors.New("This datamanager doesn't exist")},
			canProcess:          true,
			result:              result{err: "This datamanager doesn't exist"},
		},
		{
			name:                "not authorized",
			dataManagerResponse: getDataManagerResponse{dataManager: &asset.DataManager{}},
			canProcess:          false,
			result:              result{err: "not authorized to process datamanager"},
		},
		{
			name:                "not same data manager",
			dataManagerResponse: getDataManagerResponse{dataManager: &asset.DataManager{}},
			canProcess:          true,
			checkSameManagerErr: errors.New("datasamples do not share a common manager"),
			result:              result{err: "datasamples do not share a common manager"},
		},
	}

	dataManagerKey := "uuid1"
	dataSampleKeys := []string{"uuid2", "uuid3"}
	owner := "owner"

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {
				dbal := new(persistence.MockDBAL)
				ps := new(MockPermissionAPI)
				dss := new(MockDataSampleAPI)
				provider := newMockedProvider()

				dbal.On("GetDataManager", dataManagerKey).Return(c.dataManagerResponse.dataManager, c.dataManagerResponse.err)
				if c.dataManagerResponse.dataManager != nil {
					ps.On("CanProcess", c.dataManagerResponse.dataManager.Permissions, owner).Return(c.canProcess)
				}
				dss.On("CheckSameManager", dataManagerKey, dataSampleKeys).Return(c.checkSameManagerErr)

				provider.On("GetDataManagerDBAL").Return(dbal)
				provider.On("GetPermissionService").Return(ps)
				provider.On("GetDataSampleService").Return(dss)

				service := NewDataManagerService(provider)

				dmResp, err := service.GetCheckedDataManager(dataManagerKey, dataSampleKeys, owner)

				if c.result.err == "" {
					assert.Equal(t, c.result.dataManager, dmResp)
					assert.NoError(t, err)
				} else {
					assert.Nil(t, dmResp)
					assert.ErrorContains(t, err, c.result.err)
				}
			},
		)
	}
}
