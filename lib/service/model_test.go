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
	orchestrationErrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
)

func TestGetTasksModels(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	provider.On("GetModelDBAL").Return(dbal)

	service := NewModelService(provider)

	returnedModels := []*asset.Model{{}, {}, {}}

	dbal.On("GetTaskModels", "taskUuid").Once().Return(returnedModels, nil)

	models, err := service.GetTaskModels("taskUuid")
	assert.NoError(t, err)

	assert.Len(t, models, 3)
}

func TestRegisterOnNonDoingTask(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	service := NewModelService(provider)

	dbal.On("GetComputeTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(&asset.ComputeTask{
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
	assert.True(t, errors.Is(err, orchestrationErrors.ErrBadRequest))
}

func TestRegisterModelWrongPermissions(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	service := NewModelService(provider)

	dbal.On("GetComputeTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(&asset.ComputeTask{
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
	assert.True(t, errors.Is(err, orchestrationErrors.ErrPermissionDenied))
}

func TestRegisterSimpleModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)
	service := NewModelService(provider)

	dbal.On("GetComputeTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_TRAIN,
			Worker:   "test",
		},
		nil,
	)

	// No models registered
	dbal.On("GetTaskModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{}, nil)

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
}

func TestRegisterDuplicateModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetModelDBAL").Return(dbal)
	service := NewModelService(provider)

	dbal.On("GetComputeTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_TRAIN,
			Worker:   "test",
		},
		nil,
	)

	// Already one model, cannot register another one
	dbal.On("GetTaskModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{
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
	assert.True(t, errors.Is(err, orchestrationErrors.ErrBadRequest))
}

func TestRegisterHeadModel(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetModelDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)
	service := NewModelService(provider)

	dbal.On("GetComputeTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
			Worker:   "test",
		},
		nil,
	)

	// Trunk already known
	dbal.On("GetTaskModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{
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
}

func TestRegisterWrongModelType(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	service := NewModelService(provider)

	dbal.On("GetComputeTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(
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
	assert.True(t, errors.Is(err, orchestrationErrors.ErrBadRequest))
}

func TestRegisterMultipleHeads(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskDBAL").Return(dbal)
	provider.On("GetModelDBAL").Return(dbal)
	service := NewModelService(provider)

	dbal.On("GetComputeTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return(
		&asset.ComputeTask{
			Key:      "08680966-97ae-4573-8b2d-6c4db2b3c532",
			Status:   asset.ComputeTaskStatus_STATUS_DOING,
			Category: asset.ComputeTaskCategory_TASK_COMPOSITE,
			Worker:   "test",
		},
		nil,
	)

	dbal.On("GetTaskModels", "08680966-97ae-4573-8b2d-6c4db2b3c532").Once().Return([]*asset.Model{
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
	assert.True(t, errors.Is(err, orchestrationErrors.ErrBadRequest))
}