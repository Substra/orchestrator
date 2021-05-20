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
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetComputeTasksOutputModels(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	provider.On("GetModelDBAL").Return(dbal)

	service := NewModelService(provider)

	returnedModels := []*asset.Model{{}, {}, {}}

	dbal.On("GetComputeTaskOutputModels", "taskUuid").Once().Return(returnedModels, nil)

	models, err := service.GetComputeTaskOutputModels("taskUuid")
	assert.NoError(t, err)

	assert.Len(t, models, 3)

	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	provider.On("GetModelDBAL").Return(dbal)

	service := NewModelService(provider)

	model := &asset.Model{
		Key: "uuid",
	}

	dbal.On("GetModel", "uuid").Once().Return(model, nil)

	ret, err := service.GetModel("uuid")
	assert.NoError(t, err)
	assert.Equal(t, model, ret)

	provider.AssertExpectations(t)
	dbal.AssertExpectations(t)
}

func TestRegisterOnNonDoingTask(t *testing.T) {
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(&asset.ComputeTask{
		Status: asset.ComputeTaskStatus_STATUS_DONE,
		Worker: "test",
	}, nil)

	model := &asset.NewModel{
		Key:            "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:       asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.RegisterModel(model, "test")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, orcerrors.ErrBadRequest))

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterModelWrongPermissions(t *testing.T) {
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(&asset.ComputeTask{
		Status: asset.ComputeTaskStatus_STATUS_DONE,
		Worker: "owner",
	}, nil)

	model := &asset.NewModel{
		Key:            "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:       asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.RegisterModel(model, "test") // "test" is not "owner" of the task
	assert.Error(t, err)
	assert.True(t, errors.Is(err, orcerrors.ErrPermissionDenied))

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterSimpleModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	es := new(MockEventService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	service := NewModelService(provider)

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
		Worker:   "test",
		Data: &asset.ComputeTask_Train{
			Train: &asset.TrainTaskData{
				ModelPermissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				},
			},
		},
	}

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(task, nil)

	// No models registered
	dbal.On("GetComputeTaskOutputModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{}, nil)

	model := &asset.NewModel{
		Key:            "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:       asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	storedModel := &asset.Model{
		Key:            model.Key,
		Category:       model.Category,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address:        model.Address,
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
			Download: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
		},
	}
	dbal.On("AddModel", storedModel).Once().Return(nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  model.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}
	es.On("RegisterEvent", event).Once().Return(nil)

	// Model registration will initiate a task transition to done
	cts.On("applyTaskAction", task, transitionDone, mock.AnythingOfType("string")).Once().Return(nil)

	_, err := service.RegisterModel(model, "test")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestRegisterDuplicateModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	service := NewModelService(provider)

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_TRAIN,
			Worker:   "test",
		},
		nil,
	)

	// Already one model, cannot register another one
	dbal.On("GetComputeTaskOutputModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{
		{Category: asset.ModelCategory_MODEL_SIMPLE},
	}, nil)

	model := &asset.NewModel{
		Key:            "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:       asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.RegisterModel(model, "test")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, orcerrors.ErrConflict))

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterHeadModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	es := new(MockEventService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	service := NewModelService(provider)

	task := &asset.ComputeTask{
		Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
		Worker:   "test",
		Data: &asset.ComputeTask_Composite{
			Composite: &asset.CompositeTrainTaskData{
				HeadPermissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				},
				TrunkPermissions: &asset.Permissions{
					Process: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
					Download: &asset.Permission{
						Public:        true,
						AuthorizedIds: []string{},
					},
				},
			},
		},
	}
	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(task, nil)

	// Trunk already known
	dbal.On("GetComputeTaskOutputModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{
		{Category: asset.ModelCategory_MODEL_TRUNK},
	}, nil)

	model := &asset.NewModel{
		Key:            "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:       asset.ModelCategory_MODEL_HEAD,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	storedModel := &asset.Model{
		Key:            model.Key,
		Category:       model.Category,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address:        model.Address,
		Permissions: &asset.Permissions{
			Process: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
			Download: &asset.Permission{
				Public:        true,
				AuthorizedIds: []string{},
			},
		},
	}
	dbal.On("AddModel", storedModel).Once().Return(nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  model.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}
	es.On("RegisterEvent", event).Once().Return(nil)

	// Model registration will initiate a task transition to done
	cts.On("applyTaskAction", task, transitionDone, mock.AnythingOfType("string")).Once().Return(nil)

	_, err := service.RegisterModel(model, "test")
	assert.NoError(t, err)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestRegisterWrongModelType(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	service := NewModelService(provider)

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
			Worker:   "test",
		},
		nil,
	)

	dbal.On("GetComputeTaskOutputModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{}, nil)
	model := &asset.NewModel{
		Key:            "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:       asset.ModelCategory_MODEL_SIMPLE, // cannot register a SIMPLE model on composite task
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.RegisterModel(model, "test")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, orcerrors.ErrBadRequest))

	dbal.AssertExpectations(t)
	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterMultipleHeads(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	service := NewModelService(provider)

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
			Worker:   "test",
		},
		nil,
	)

	dbal.On("GetComputeTaskOutputModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{
		{Category: asset.ModelCategory_MODEL_HEAD},
	}, nil)

	model := &asset.NewModel{
		Key:            "18680966-97ae-4573-8b2d-6c4db2b3c532",
		Category:       asset.ModelCategory_MODEL_HEAD,
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		Address: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	_, err := service.RegisterModel(model, "test")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, orcerrors.ErrConflict))

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetInputModels(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("GetTask", "uuid").Once().Return(
		&asset.ComputeTask{
			ParentTaskKeys: []string{"parent1", "parent2"},
		},
		nil,
	)

	model1 := &asset.Model{Key: "m1"}
	model2 := &asset.Model{Key: "m2"}

	dbal.On("GetComputeTaskOutputModels", "parent1").Once().Return([]*asset.Model{model1}, nil)
	dbal.On("GetComputeTaskOutputModels", "parent2").Once().Return([]*asset.Model{model2}, nil)

	models, err := service.GetComputeTaskInputModels("uuid")
	assert.NoError(t, err)

	assert.Equal(t, []*asset.Model{model1, model2}, models)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestCanDisableModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewModelService(provider)

	cts.On("canDisableModels", "taskKey", "requester").Once().Return(true, nil)

	dbal.On("GetModel", "modelUuid").Once().Return(&asset.Model{
		Key:            "modelUuid",
		ComputeTaskKey: "taskKey",
	}, nil)

	can, err := service.CanDisableModel("modelUuid", "requester")
	assert.NoError(t, err)
	assert.True(t, can)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestDisableModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	es := new(MockEventService)
	provider := new(MockServiceProvider)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetEventService").Return(es)
	service := NewModelService(provider)

	cts.On("canDisableModels", "taskKey", "requester").Return(true, nil)

	dbal.On("GetModel", "modelUuid").Return(&asset.Model{
		Key:            "modelUuid",
		ComputeTaskKey: "taskKey",
		Address:        &asset.Addressable{Checksum: "sha", StorageAddress: "http://there"},
	}, nil)

	dbal.On("UpdateModel", &asset.Model{Key: "modelUuid", ComputeTaskKey: "taskKey"}).Return(nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_MODEL,
		AssetKey:  "modelUuid",
		EventKind: asset.EventKind_EVENT_ASSET_DISABLED,
	}
	es.On("RegisterEvent", event).Once().Return(nil)

	err := service.DisableModel("modelUuid", "requester")
	assert.NoError(t, err)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestQueryModels(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetModelDBAL").Return(dbal)
	service := NewModelService(provider)

	model1 := asset.Model{
		Key:      "model1",
		Category: asset.ModelCategory_MODEL_SIMPLE,
	}
	model2 := asset.Model{
		Key:      "model2",
		Category: asset.ModelCategory_MODEL_SIMPLE,
	}

	pagination := common.NewPagination("", 12)

	dbal.On("QueryModels", asset.ModelCategory_MODEL_SIMPLE, pagination).Return([]*asset.Model{&model1, &model2}, "nextPage", nil).Once()

	r, token, err := service.QueryModels(asset.ModelCategory_MODEL_SIMPLE, pagination)
	require.Nil(t, err)

	assert.Len(t, r, 2)
	assert.Equal(t, r[0].Key, model1.Key)
	assert.Equal(t, "nextPage", token, "next page token should be returned")
}

func TestAreAllOutputsRegistered(t *testing.T) {
	cases := map[string]struct {
		task    *asset.ComputeTask
		models  []*asset.Model
		outcome bool
	}{
		"train without model": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_TRAIN},
			models:  []*asset.Model{},
			outcome: false,
		},
		"unhandled task category": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_TEST},
			models:  []*asset.Model{},
			outcome: false,
		},
		"train with model": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_TRAIN},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_SIMPLE}},
			outcome: true,
		},
		"aggregate with model": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_AGGREGATE},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_SIMPLE}},
			outcome: true,
		},
		"composite with head": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_COMPOSITE},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_HEAD}},
			outcome: false,
		},
		"composite with trunk": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_COMPOSITE},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_TRUNK}},
			outcome: false,
		},
		"composite with head & trunk": {
			task:    &asset.ComputeTask{Category: asset.ComputeTaskCategory_TASK_COMPOSITE},
			models:  []*asset.Model{{Category: asset.ModelCategory_MODEL_TRUNK}, {Category: asset.ModelCategory_MODEL_HEAD}},
			outcome: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, areAllOutputsRegistered(tc.task, tc.models))
		})
	}
}

func TestCheckDuplicateModel(t *testing.T) {
	cases := map[string]struct {
		models      []*asset.Model
		model       *asset.NewModel
		shouldError bool
	}{
		"no models": {
			[]*asset.Model{},
			&asset.NewModel{Category: asset.ModelCategory_MODEL_SIMPLE},
			false,
		},
		"simple": {
			[]*asset.Model{{Category: asset.ModelCategory_MODEL_SIMPLE}},
			&asset.NewModel{Category: asset.ModelCategory_MODEL_SIMPLE},
			true,
		},
		"head": {
			[]*asset.Model{{Category: asset.ModelCategory_MODEL_HEAD}},
			&asset.NewModel{Category: asset.ModelCategory_MODEL_HEAD},
			true,
		},
		"head and trunk": {
			[]*asset.Model{{Category: asset.ModelCategory_MODEL_HEAD}},
			&asset.NewModel{Category: asset.ModelCategory_MODEL_TRUNK},
			false,
		},
		"complete composite": {
			[]*asset.Model{{Category: asset.ModelCategory_MODEL_HEAD}, {Category: asset.ModelCategory_MODEL_TRUNK}},
			&asset.NewModel{Category: asset.ModelCategory_MODEL_TRUNK},
			true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tc.shouldError {
				assert.Error(t, checkDuplicateModel(tc.models, tc.model))
			} else {
				assert.NoError(t, checkDuplicateModel(tc.models, tc.model))
			}
		})
	}
}
