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
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
)

type ModelAPI interface {
	RegisterModel(model *asset.NewModel, owner string) (*asset.Model, error)
	GetTaskModels(key string) ([]*asset.Model, error)
}

type ModelServiceProvider interface {
	GetModelService() ModelAPI
}

type ModelDependencyProvider interface {
	persistence.ModelDBALProvider
	persistence.ComputeTaskDBALProvider
	persistence.AlgoDBALProvider
	persistence.DataManagerDBALProvider
	PermissionServiceProvider
	event.QueueProvider
}

type ModelService struct {
	ModelDependencyProvider
}

func NewModelService(provider ModelDependencyProvider) *ModelService {
	return &ModelService{provider}
}

func (s *ModelService) GetTaskModels(key string) ([]*asset.Model, error) {
	return s.GetModelDBAL().GetTaskModels(key)
}

func (s *ModelService) RegisterModel(newModel *asset.NewModel, requester string) (*asset.Model, error) {
	log.WithField("model", newModel).WithField("requester", requester).Debug("Registering new model")

	err := newModel.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrInvalidAsset, err.Error())
	}

	task, err := s.GetComputeTaskDBAL().GetComputeTask(newModel.ComputeTaskKey)
	if err != nil {
		return nil, err
	}

	if task.Worker != requester {
		return nil, fmt.Errorf("%w: only \"%s\" worker can register model", errors.ErrPermissionDenied, task.Worker)
	}

	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		return nil, fmt.Errorf("%w: cannot register model for taks with status \"%s\"", errors.ErrBadRequest, task.Status.String())
	}

	var model *asset.Model

	switch task.Category {
	case asset.ComputeTaskCategory_TASK_AGGREGATE, asset.ComputeTaskCategory_TASK_TRAIN:
		model, err = s.registerSimpleModel(newModel, requester, task)
		if err != nil {
			return nil, err
		}
	case asset.ComputeTaskCategory_TASK_COMPOSITE:
		model, err = s.registerCompositeModel(newModel, requester, task)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unhandled model category")
	}

	err = s.GetModelDBAL().AddModel(model)
	if err != nil {
		return nil, err
	}

	event := &event.Event{
		AssetKind: asset.ModelKind,
		AssetID:   model.Key,
		EventKind: event.AssetCreated,
	}

	err = s.GetEventQueue().Enqueue(event)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func (s *ModelService) registerSimpleModel(newModel *asset.NewModel, requester string, task *asset.ComputeTask) (*asset.Model, error) {
	// This should be checked by caller, but better safe than sorry
	if !(task.Category == asset.ComputeTaskCategory_TASK_TRAIN || task.Category == asset.ComputeTaskCategory_TASK_AGGREGATE) {
		return nil, fmt.Errorf("%w: cannot register train model for \"%s\" task", errors.ErrBadRequest, task.Category.String())
	}
	if newModel.Category != asset.ModelCategory_MODEL_SIMPLE {
		return nil, fmt.Errorf("%w: cannot register non-simple model", errors.ErrBadRequest)
	}
	existingModels, err := s.GetModelDBAL().GetTaskModels(task.Key)
	if err != nil {
		return nil, err
	}
	if len(existingModels) > 0 {
		return nil, fmt.Errorf("%w: task already has a model: %s", errors.ErrBadRequest, existingModels[0].Key)
	}

	model := &asset.Model{
		Key:            newModel.Key,
		Category:       newModel.Category,
		ComputeTaskKey: task.Key,
		Address:        newModel.Address,
	}

	return model, nil
}

func (s *ModelService) registerCompositeModel(newModel *asset.NewModel, requester string, task *asset.ComputeTask) (*asset.Model, error) {
	// This should be checked by caller, but better safe than sorry
	if task.Category != asset.ComputeTaskCategory_TASK_COMPOSITE {
		return nil, fmt.Errorf("%w: cannot register composite model for \"%s\" task", errors.ErrBadRequest, task.Category.String())
	}
	if !(newModel.Category == asset.ModelCategory_MODEL_HEAD || newModel.Category == asset.ModelCategory_MODEL_TRUNK) {
		return nil, fmt.Errorf("%w: cannot register non-composite model", errors.ErrBadRequest)
	}
	existingModels, err := s.GetModelDBAL().GetTaskModels(task.Key)
	if err != nil {
		return nil, err
	}
	for _, m := range existingModels {
		if m.Category == newModel.Category {
			return nil, fmt.Errorf("%w: task already has a \"%s\" model: %s", errors.ErrBadRequest, newModel.Category.String(), m.Key)
		}
	}

	model := &asset.Model{
		Key:            newModel.Key,
		Category:       newModel.Category,
		ComputeTaskKey: task.Key,
		Address:        newModel.Address,
	}

	return model, nil
}
