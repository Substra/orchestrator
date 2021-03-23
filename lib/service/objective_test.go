// Copyright 2020 Owkin Inc.
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
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/event"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterObjective(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)

	provider.On("GetObjectiveDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)

	provider.On("GetEventQueue").Return(dispatcher)

	service := NewObjectiveService(provider)

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	metrics := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	newPerms := &asset.NewPermissions{Public: true}

	objective := &asset.NewObjective{
		Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:           "Test objective",
		MetricsName:    "test perf",
		Metrics:        metrics,
		Description:    description,
		NewPermissions: newPerms,
	}

	e := &event.Event{
		EventKind: event.AssetCreated,
		AssetKind: asset.ObjectiveKind,
		AssetID:   objective.Key,
	}
	dispatcher.On("Enqueue", e).Return(nil)

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	storedObjective := &asset.Objective{
		Key:         "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:        "Test objective",
		MetricsName: "test perf",
		Metrics:     metrics,
		Description: description,
		Permissions: perms,
		Owner:       "owner",
	}

	mps.On("CreatePermissions", "owner", newPerms).Return(perms, nil).Once()
	dbal.On(
		"AddObjective",
		storedObjective,
	).Return(nil).Once()

	o, err := service.RegisterObjective(objective, "owner")

	assert.NoError(t, err, "Registration of valid objective should not fail")
	assert.NotNil(t, o, "Registration should return an Objective")
	assert.Equal(t, perms, o.Permissions, "Permissions should be set")
	assert.Equal(t, "owner", o.Owner, "Owner should be set")

	dbal.AssertExpectations(t)
}

func TestGetObjective(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetObjectiveDBAL").Return(dbal)
	service := NewObjectiveService(provider)

	objective := asset.Objective{
		Key:  "objKey",
		Name: "Test",
	}

	dbal.On("GetObjective", "objKey").Return(&objective, nil).Once()

	o, err := service.GetObjective("objKey")
	require.Nil(t, err)
	assert.Equal(t, o.Name, objective.Name)
}

func TestGetObjectives(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetObjectiveDBAL").Return(dbal)
	service := NewObjectiveService(provider)

	obj1 := asset.Objective{
		Key:  "obj1",
		Name: "Test 1",
	}
	obj2 := asset.Objective{
		Key:  "obj2",
		Name: "Test 2",
	}

	pagination := common.NewPagination("", 12)

	dbal.On("GetObjectives", pagination).Return([]*asset.Objective{&obj1, &obj2}, "nextPage", nil).Once()

	r, token, err := service.GetObjectives(pagination)
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, obj1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}

func TestObjectiveExists(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetObjectiveDBAL").Return(dbal)
	service := NewObjectiveService(provider)

	dbal.On("ObjectiveExists", "obj1").Return(true, nil).Once()

	ok, err := service.ObjectiveExists("obj1")

	assert.Equal(t, ok, true)
	assert.NoError(t, err)
}

func TestCanDownload(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetObjectiveDBAL").Return(dbal)
	service := NewObjectiveService(provider)

	perms := &asset.Permissions{
		Process: &asset.Permission{Public: true},
		Download: &asset.Permission{
			Public:        false,
			AuthorizedIds: []string{"org-2"},
		},
	}

	objective := &asset.Objective{
		Key:         "obj1",
		Name:        "Test",
		Permissions: perms,
	}

	dbal.On("GetObjective", "obj1").Return(objective, nil).Once()

	ok, err := service.CanDownload("obj1", "org-2")

	assert.Equal(t, ok, true)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
}
