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
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
)

type ModelAPI interface {
	RegisterModel(model *asset.NewModel, owner string) (*asset.Model, error)
	GetComputeTaskOutputModels(key string) ([]*asset.Model, error)
	GetComputeTaskInputModels(key string) ([]*asset.Model, error)
	CanDisableModel(key, requester string) (bool, error)
	// DisableModel removes a model address and emit a "disabled" event
	DisableModel(key string, requester string) error
	GetModel(key string) (*asset.Model, error)
	QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error)
}

type ModelServiceProvider interface {
	GetModelService() ModelAPI
}

type ModelDependencyProvider interface {
	persistence.ModelDBALProvider
	persistence.AlgoDBALProvider
	persistence.DataManagerDBALProvider
	PermissionServiceProvider
	ComputeTaskServiceProvider
	ComputePlanServiceProvider
	event.QueueProvider
}

type ModelService struct {
	ModelDependencyProvider
}

func NewModelService(provider ModelDependencyProvider) *ModelService {
	return &ModelService{provider}
}

func (s *ModelService) GetComputeTaskOutputModels(key string) ([]*asset.Model, error) {
	return s.GetModelDBAL().GetComputeTaskOutputModels(key)
}

func (s *ModelService) GetModel(key string) (*asset.Model, error) {
	log.WithField("key", key).Debug("Get model")
	return s.GetModelDBAL().GetModel(key)
}

func (s *ModelService) QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error) {
	return s.GetModelDBAL().QueryModels(c, p)
}

// GetComputeTaskInputModels retrieves input models of a given task from its parents.
func (s *ModelService) GetComputeTaskInputModels(key string) ([]*asset.Model, error) {
	task, err := s.GetComputeTaskService().GetTask(key)
	if err != nil {
		return nil, err
	}

	inputs := []*asset.Model{}

	for _, p := range task.ParentTaskKeys {
		models, err := s.GetComputeTaskOutputModels(p)
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, models...)
	}

	return inputs, nil
}

func (s *ModelService) RegisterModel(newModel *asset.NewModel, requester string) (*asset.Model, error) {
	log.WithField("model", newModel).WithField("requester", requester).Debug("Registering new model")

	err := newModel.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrInvalidAsset, err.Error())
	}

	task, err := s.GetComputeTaskService().GetTask(newModel.ComputeTaskKey)
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
		return nil, fmt.Errorf("unhandled model category")
	}

	err = s.GetModelDBAL().AddModel(model)
	if err != nil {
		return nil, err
	}

	event := event.NewEvent(event.AssetCreated, model.Key, asset.ModelKind)

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
	existingModels, err := s.GetModelDBAL().GetComputeTaskOutputModels(task.Key)
	if err != nil {
		return nil, err
	}
	if len(existingModels) > 0 {
		return nil, fmt.Errorf("%w: task already has a model: %s", errors.ErrBadRequest, existingModels[0].Key)
	}

	var permissions *asset.Permissions

	switch task.Category {
	case asset.ComputeTaskCategory_TASK_TRAIN:
		permissions = task.GetTrain().ModelPermissions
	case asset.ComputeTaskCategory_TASK_AGGREGATE:
		permissions = task.GetAggregate().ModelPermissions
	default:
		return nil, fmt.Errorf("%w: cannot set model permissions for \"%s\" task", errors.ErrBadRequest, task.Category.String())
	}

	model := &asset.Model{
		Key:            newModel.Key,
		Category:       newModel.Category,
		ComputeTaskKey: task.Key,
		Address:        newModel.Address,
		Permissions:    permissions,
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
	existingModels, err := s.GetModelDBAL().GetComputeTaskOutputModels(task.Key)
	if err != nil {
		return nil, err
	}
	for _, m := range existingModels {
		if m.Category == newModel.Category {
			return nil, fmt.Errorf("%w: task already has a \"%s\" model: %s", errors.ErrBadRequest, newModel.Category.String(), m.Key)
		}
	}

	var permissions *asset.Permissions

	switch newModel.Category {
	case asset.ModelCategory_MODEL_HEAD:
		permissions = task.GetComposite().HeadPermissions
	case asset.ModelCategory_MODEL_TRUNK:
		permissions = task.GetComposite().TrunkPermissions
	default:
		return nil, fmt.Errorf("%w: cannot set permissions for \"%s\" model", errors.ErrBadRequest, newModel.Category.String())
	}

	model := &asset.Model{
		Key:            newModel.Key,
		Category:       newModel.Category,
		ComputeTaskKey: task.Key,
		Address:        newModel.Address,
		Permissions:    permissions,
	}

	return model, nil
}

// CanDisableModel returns true if the model can be disabled
func (s *ModelService) CanDisableModel(key string, requester string) (bool, error) {
	model, err := s.GetModelDBAL().GetModel(key)
	if err != nil {
		return false, err
	}

	return s.canDisableModel(model, requester)
}

func (s *ModelService) canDisableModel(model *asset.Model, requester string) (bool, error) {
	return s.GetComputeTaskService().canDisableModels(model.ComputeTaskKey, requester)
}

// DisableModel removes model's address and emit an "disabled" event
func (s *ModelService) DisableModel(key string, requester string) error {
	log.WithField("modelKey", key).Debug("disabling model")
	model, err := s.GetModelDBAL().GetModel(key)
	if err != nil {
		return err
	}

	canClean, err := s.canDisableModel(model, requester)
	if err != nil {
		return err
	}
	if !canClean {
		return fmt.Errorf("cannot disable a model in use: %w", errors.ErrCannotDisableModel)
	}

	model.Address = nil

	err = s.GetModelDBAL().UpdateModel(model)
	if err != nil {
		return err
	}

	event := event.NewEvent(event.AssetDisabled, model.Key, asset.ModelKind)
	err = s.GetEventQueue().Enqueue(event)
	if err != nil {
		return err
	}

	return nil
}
