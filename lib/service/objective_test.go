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

func TestRegisterObjective(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	es := new(MockEventService)
	provider := new(MockServiceProvider)

	provider.On("GetObjectiveDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)

	provider.On("GetEventService").Return(es)

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

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_OBJECTIVE,
		AssetKey:  objective.Key,
	}
	es.On("RegisterEvents", []*asset.Event{e}).Once().Return(nil)

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

func TestQueryObjectives(t *testing.T) {
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

	dbal.On("QueryObjectives", pagination).Return([]*asset.Objective{&obj1, &obj2}, "nextPage", nil).Once()

	r, token, err := service.QueryObjectives(pagination)
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
		Key:         "837B2E87-35CA-48F9-B83C-B40FB3FBA4E6",
		Name:        "Test",
		Permissions: perms,
	}

	dbal.On("GetObjective", "obj1").Return(objective, nil).Once()

	ok, err := service.CanDownload("obj1", "org-2")

	assert.Equal(t, ok, true)
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
}

func TestRegisterObjectiveWithDatamanager(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mps := new(MockPermissionService)
	mds := new(MockDataSampleService)
	mdm := new(MockDataManagerService)
	es := new(MockEventService)
	provider := new(MockServiceProvider)

	provider.On("GetObjectiveDBAL").Return(dbal)
	provider.On("GetPermissionService").Return(mps)
	provider.On("GetDataSampleService").Return(mds)
	provider.On("GetDataManagerService").Return(mdm)
	provider.On("GetEventService").Return(es)

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
		DataManagerKey: "34a251cc-23f0-456f-95f3-0952524f718b",
		DataSampleKeys: []string{"6c34f9da-5575-44f6-8f02-d911d3898f77"},
	}

	dataManagerUpdate := &asset.DataManagerUpdateParam{
		Key:          objective.DataManagerKey,
		ObjectiveKey: objective.Key,
	}

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_OBJECTIVE,
		AssetKey:  objective.Key,
	}
	es.On("RegisterEvents", []*asset.Event{e}).Once().Return(nil)

	mds.On("CheckSameManager", objective.DataManagerKey, objective.DataSampleKeys).Return(nil).Once()
	mds.On("IsTestOnly", objective.DataSampleKeys).Return(true, nil).Once()
	mdm.On("UpdateDataManager", dataManagerUpdate, "owner").Return(nil).Once()

	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	storedObjective := &asset.Objective{
		Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:           "Test objective",
		MetricsName:    "test perf",
		Metrics:        metrics,
		Description:    description,
		Permissions:    perms,
		Owner:          "owner",
		DataManagerKey: objective.DataManagerKey,
		DataSampleKeys: objective.DataSampleKeys,
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

func TestRejectInvalidDatamanager(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	mds := new(MockDataSampleService)
	provider := new(MockServiceProvider)

	provider.On("GetObjectiveDBAL").Return(dbal)
	provider.On("GetDataSampleService").Return(mds)

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
		DataManagerKey: "34a251cc-23f0-456f-95f3-0952524f718b",
		DataSampleKeys: []string{"6c34f9da-5575-44f6-8f02-d911d3898f77"},
	}

	mds.On("CheckSameManager", objective.DataManagerKey, objective.DataSampleKeys).Return(errors.New("not the same datamanager")).Once()

	_, err := service.RegisterObjective(objective, "owner")

	assert.Error(t, err, "Registration should fail")
}

func TestGetLeaderBoard(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	obs := new(MockObjectiveService)
	provider := new(MockServiceProvider)

	provider.On("GetObjectiveDBAL").Return(dbal)
	provider.On("GetObjectiveService").Return(obs)

	service := NewObjectiveService(provider)

	objective := &asset.Objective{
		Key:            "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:           "Test objective",
		DataManagerKey: "34a251cc-23f0-456f-95f3-0952524f718b",
		DataSampleKeys: []string{"6c34f9da-5575-44f6-8f02-d911d3898f77"},
	}

	description := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	algorithm := &asset.Addressable{
		StorageAddress: "ftp://127.0.0.1/test",
		Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
	}
	perms := &asset.Permissions{Process: &asset.Permission{Public: true}}

	algo := &asset.Algo{
		Key:         "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Name:        "Test algo",
		Category:    asset.AlgoCategory_ALGO_SIMPLE,
		Algorithm:   algorithm,
		Description: description,
		Permissions: perms,
		Owner:       "owner",
	}

	BoardItem := &asset.BoardItem{
		Algo:           algo,
		ObjectiveKey:   "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskKey: "867852b4-8419-4d52-8862-d5db823095be",
		Perf:           0.36492,
	}

	leaderboard := &asset.Leaderboard{
		Objective:  objective,
		BoardItems: []*asset.BoardItem{BoardItem},
	}

	dbal.On("GetLeaderboard", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(leaderboard, nil).Once()

	o, err := service.GetLeaderboard(&asset.LeaderboardQueryParam{
		ObjectiveKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		SortOrder:    asset.SortOrder_ASCENDING,
	})

	assert.Equal(t, o.BoardItems, []*asset.BoardItem{BoardItem})
	dbal.AssertExpectations(t)
	require.Nil(t, err)
}
