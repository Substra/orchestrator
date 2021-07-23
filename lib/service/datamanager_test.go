package service

import (
	"errors"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDataManager(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	mps := new(MockPermissionAPI)
	obj := new(MockObjectiveAPI)
	provider := new(MockDependenciesProvider)
	es := new(MockEventAPI)

	provider.On("GetObjectiveService").Return(obj)
	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetEventService").Return(es)

	service := NewDataManagerService(provider)

	newPerms := &asset.NewPermissions{Public: true}

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

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
		ObjectiveKey:   "da9b3341-0539-44cb-835d-0baeb5644151",
		Description:    description,
		Opener:         opener,
		Type:           "test dm",
		NewPermissions: newPerms,
	}

	storedDataManager := &asset.DataManager{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:         "test datamanager",
		Owner:        "owner",
		Permissions:  perms,
		ObjectiveKey: "da9b3341-0539-44cb-835d-0baeb5644151",
		Description:  description,
		Opener:       opener,
		Type:         "test dm",
	}

	mps.On("CreatePermissions", "owner", newPerms).Return(perms, nil).Once()
	dbal.On("DataManagerExists", newDataManager.GetKey()).Return(false, nil).Once()
	dbal.On("AddDataManager", storedDataManager).Return(nil).Once()
	obj.On("CanDownload", "da9b3341-0539-44cb-835d-0baeb5644151", "owner").Return(true, nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
		AssetKey:  storedDataManager.Key,
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	dm, err := service.RegisterDataManager(newDataManager, "owner")

	assert.NoError(t, err, "Registration of valid datamanager should not fail")
	assert.NotNil(t, dm, "Registratrion should return a datamanager asset")
	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestRegisterDataManagerEmptyObjective(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	mps := new(MockPermissionAPI)
	provider := new(MockDependenciesProvider)
	es := new(MockEventAPI)

	provider.On("GetObjectiveDBAL").Return(dbal)
	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetEventService").Return(es)

	service := NewDataManagerService(provider)

	newPerms := &asset.NewPermissions{Public: true}

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

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
		ObjectiveKey:   "",
		Description:    description,
		Opener:         opener,
		Type:           "test dm",
		NewPermissions: newPerms,
	}

	storedDataManager := &asset.DataManager{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:         "test datamanager",
		Owner:        "owner",
		Permissions:  perms,
		ObjectiveKey: "",
		Description:  description,
		Opener:       opener,
		Type:         "test dm",
	}

	mps.On("CreatePermissions", "owner", newPerms).Return(perms, nil).Once()
	dbal.On("DataManagerExists", newDataManager.GetKey()).Return(false, nil).Once()
	dbal.On("AddDataManager", storedDataManager).Return(nil).Once()

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
		AssetKey:  storedDataManager.Key,
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	dm, err := service.RegisterDataManager(newDataManager, "owner")

	assert.NoError(t, err, "Registration of valid datamanager should not fail")
	assert.NotNil(t, dm, "Registration should return a data manager")
	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestRegisterDataManagerUnknownObjective(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	mps := new(MockPermissionAPI)
	obj := new(MockObjectiveAPI)
	provider := new(MockDependenciesProvider)

	provider.On("GetObjectiveService").Return(obj)
	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)

	service := NewDataManagerService(provider)

	newPerms := &asset.NewPermissions{Public: true}

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
		ObjectiveKey:   "da9b3341-0539-44cb-835d-0baeb5644151",
		Description:    description,
		Opener:         opener,
		Type:           "test dm",
		NewPermissions: newPerms,
	}

	dbal.On("DataManagerExists", newDataManager.GetKey()).Return(false, nil).Once()
	obj.On("CanDownload", "da9b3341-0539-44cb-835d-0baeb5644151", "owner").Return(false, errors.New("not found")).Once()

	_, err := service.RegisterDataManager(newDataManager, "owner")

	assert.Error(t, err, "Registration of an invalid datamanager should fail")
	obj.AssertExpectations(t)
}

func TestUpdateDataManager(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	mps := new(MockPermissionAPI)
	obj := new(MockObjectiveAPI)
	provider := new(MockDependenciesProvider)
	es := new(MockEventAPI)

	provider.On("GetObjectiveService").Return(obj)
	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetEventService").Return(es)

	service := NewDataManagerService(provider)

	dataManagerUpdate := &asset.DataManagerUpdateParam{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		ObjectiveKey: "da9b3341-0539-44cb-835d-0baeb5644151",
	}

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	opener := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	storedDataManager := &asset.DataManager{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:         "test datamanager",
		Owner:        "owner",
		Permissions:  perms,
		ObjectiveKey: "",
		Description:  description,
		Opener:       opener,
		Type:         "test dm",
	}

	updatedDataManager := &asset.DataManager{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:         "test datamanager",
		Owner:        "owner",
		Permissions:  perms,
		ObjectiveKey: "da9b3341-0539-44cb-835d-0baeb5644151",
		Description:  description,
		Opener:       opener,
		Type:         "test dm",
	}

	dbal.On("GetDataManager", "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6").Return(storedDataManager, nil).Once()
	obj.On("CanDownload", "da9b3341-0539-44cb-835d-0baeb5644151", "owner").Return(true, nil).Once()
	dbal.On("UpdateDataManager", updatedDataManager).Return(nil).Once()
	mps.On("CanProcess", perms, "owner").Return(true)

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
		AssetKey:  updatedDataManager.Key,
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	err := service.UpdateDataManager(dataManagerUpdate, "owner")

	assert.NoError(t, err, "Update should not fail")
	dbal.AssertExpectations(t)
	obj.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateDataManagerOtherOwner(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	mps := new(MockPermissionAPI)
	provider := new(MockDependenciesProvider)
	obj := new(MockObjectiveAPI)
	es := new(MockEventAPI)

	provider.On("GetObjectiveService").Return(obj)
	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetEventService").Return(es)

	service := NewDataManagerService(provider)

	dataManagerUpdate := &asset.DataManagerUpdateParam{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		ObjectiveKey: "da9b3341-0539-44cb-835d-0baeb5644151",
	}

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	opener := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	storedDataManager := &asset.DataManager{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:         "test datamanager",
		Owner:        "owner",
		Permissions:  perms,
		ObjectiveKey: "",
		Description:  description,
		Opener:       opener,
		Type:         "test dm",
	}

	updatedDataManager := &asset.DataManager{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:         "test datamanager",
		Owner:        "owner",
		Permissions:  perms,
		ObjectiveKey: "da9b3341-0539-44cb-835d-0baeb5644151",
		Description:  description,
		Opener:       opener,
		Type:         "test dm",
	}

	dbal.On("GetDataManager", "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6").Return(storedDataManager, nil).Once()
	obj.On("CanDownload", "da9b3341-0539-44cb-835d-0baeb5644151", "owner").Return(true, nil).Once()
	dbal.On("UpdateDataManager", updatedDataManager).Return(nil).Once()
	mps.On("CanProcess", perms, "other_owner").Return(true)

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
		AssetKey:  updatedDataManager.Key,
	}
	es.On("RegisterEvents", e).Once().Return(nil)

	err := service.UpdateDataManager(dataManagerUpdate, "other_owner")

	assert.NoError(t, err, "Update should not fail")
	dbal.AssertExpectations(t)
	obj.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestUpdateDataManagerObjectiveKeyAlreadySet(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	mps := new(MockPermissionAPI)
	provider := new(MockDependenciesProvider)

	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)

	service := NewDataManagerService(provider)

	dataManagerUpdate := &asset.DataManagerUpdateParam{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		ObjectiveKey: "da9b3341-0539-44cb-835d-0baeb5644151",
	}

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	opener := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	storedDataManager := &asset.DataManager{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:         "test datamanager",
		Owner:        "owner",
		Permissions:  perms,
		ObjectiveKey: "3e8e3f29-c48e-4d56-a709-14e108fadccf",
		Description:  description,
		Opener:       opener,
		Type:         "test dm",
	}

	dbal.On("GetDataManager", "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6").Return(storedDataManager, nil).Once()
	mps.On("CanProcess", perms, "owner").Return(true)

	err := service.UpdateDataManager(dataManagerUpdate, "owner")

	assert.Error(t, err, "Update should fail")
	dbal.AssertExpectations(t)
}

func TestUpdateDataManagerUnknownObjective(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	mps := new(MockPermissionAPI)
	provider := new(MockDependenciesProvider)
	obj := new(MockObjectiveAPI)

	provider.On("GetObjectiveService").Return(obj)
	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)

	service := NewDataManagerService(provider)

	dataManagerUpdate := &asset.DataManagerUpdateParam{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		ObjectiveKey: "da9b3341-0539-44cb-835d-0baeb5644151",
	}

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	opener := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	storedDataManager := &asset.DataManager{
		Key:          "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6",
		Name:         "test datamanager",
		Owner:        "owner",
		Permissions:  perms,
		ObjectiveKey: "",
		Description:  description,
		Opener:       opener,
		Type:         "test dm",
	}

	dbal.On("GetDataManager", "65afb5fe-f6bc-4f8c-b488-f5e24a9d94a6").Return(storedDataManager, nil).Once()
	obj.On("CanDownload", "da9b3341-0539-44cb-835d-0baeb5644151", "owner").Return(false, errors.New("unknown objective")).Once()
	mps.On("CanProcess", perms, "owner").Return(true)

	err := service.UpdateDataManager(dataManagerUpdate, "owner")

	assert.Error(t, err, "Update should fail")
	dbal.AssertExpectations(t)
	obj.AssertExpectations(t)
}

func TestGetDataManager(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	provider := new(MockDependenciesProvider)
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
	dbal := new(persistenceHelper.DBAL)
	provider := new(MockDependenciesProvider)
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
	dbal := new(persistenceHelper.DBAL)
	provider := new(MockDependenciesProvider)
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
