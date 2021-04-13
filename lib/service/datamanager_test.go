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

package service

import (
	"errors"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDataManager(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	obj := new(MockObjectiveService)
	provider := new(MockServiceProvider)

	provider.On("GetObjectiveService").Return(obj)
	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)

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

	err := service.RegisterDataManager(newDataManager, "owner")

	assert.NoError(t, err, "Registration of valid datamanager should not fail")
	dbal.AssertExpectations(t)
}

func TestRegisterDataManagerEmptyObjective(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	provider := new(MockServiceProvider)

	provider.On("GetObjectiveDBAL").Return(dbal)
	provider.On("GetDataManagerDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)

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

	err := service.RegisterDataManager(newDataManager, "owner")

	assert.NoError(t, err, "Registration of valid datamanager should not fail")
	dbal.AssertExpectations(t)
}

func TestRegisterDataManagerUnknownObjective(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	obj := new(MockObjectiveService)
	provider := new(MockServiceProvider)

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

	err := service.RegisterDataManager(newDataManager, "owner")

	assert.Error(t, err, "Registration of valid datamanager should not fail")
	obj.AssertExpectations(t)
}

func TestUpdateDataManager(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	obj := new(MockObjectiveService)
	provider := new(MockServiceProvider)

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

	err := service.UpdateDataManager(dataManagerUpdate, "owner")

	assert.NoError(t, err, "Update should not fail")
	dbal.AssertExpectations(t)
	obj.AssertExpectations(t)
}

func TestUpdateDataManagerOtherOwner(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	provider := new(MockServiceProvider)
	obj := new(MockObjectiveService)

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

	err := service.UpdateDataManager(dataManagerUpdate, "other_owner")

	assert.NoError(t, err, "Update should not fail")
	dbal.AssertExpectations(t)
	obj.AssertExpectations(t)
}

func TestUpdateDataManagerObjectiveKeyAlreadySet(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	provider := new(MockServiceProvider)

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
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	provider := new(MockServiceProvider)
	obj := new(MockObjectiveService)

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
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
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

func TestGetDataManagers(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
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

	dbal.On("GetDataManagers", pagination).Return([]*asset.DataManager{&dm1, &dm2}, "nextPage", nil).Once()

	r, token, err := service.GetDataManagers(pagination)
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, dm1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
	dbal.AssertExpectations(t)
}

func TestIsOwner(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
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
