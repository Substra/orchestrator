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
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
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
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)
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
	}
	dbal.On("AddModel", storedModel).Once().Return(nil)

	event := &event.Event{
		AssetKind: asset.ModelKind,
		AssetID:   model.Key,
		EventKind: event.AssetCreated,
	}
	dispatcher.On("Enqueue", event).Once().Return(nil)

	_, err := service.RegisterModel(model, "test")
	assert.NoError(t, err)

	dbal.AssertExpectations(t)
	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
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
	assert.True(t, errors.Is(err, orcerrors.ErrBadRequest))

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterHeadModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)
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
	}
	dbal.On("AddModel", storedModel).Once().Return(nil)

	event := &event.Event{
		AssetKind: asset.ModelKind,
		AssetID:   model.Key,
		EventKind: event.AssetCreated,
	}
	dispatcher.On("Enqueue", event).Once().Return(nil)

	_, err := service.RegisterModel(model, "test")
	assert.NoError(t, err)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestRegisterWrongModelType(t *testing.T) {
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
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
	assert.True(t, errors.Is(err, orcerrors.ErrBadRequest))

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
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetEventQueue").Return(dispatcher)
	service := NewModelService(provider)

	cts.On("canDisableModels", "taskKey", "requester").Return(true, nil)

	dbal.On("GetModel", "modelUuid").Return(&asset.Model{
		Key:            "modelUuid",
		ComputeTaskKey: "taskKey",
		Address:        &asset.Addressable{Checksum: "sha", StorageAddress: "http://there"},
	}, nil)

	dbal.On("UpdateModel", &asset.Model{Key: "modelUuid", ComputeTaskKey: "taskKey"}).Return(nil)

	dispatcher.On("Enqueue", &event.Event{
		AssetKind: asset.ModelKind,
		AssetID:   "modelUuid",
		EventKind: event.AssetDisabled,
	}).Return(nil)

	err := service.DisableModel("modelUuid", "requester")
	assert.NoError(t, err)

	cts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	provider.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}
