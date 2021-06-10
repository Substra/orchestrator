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
	EventServiceProvider
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

		if task.Category == asset.ComputeTaskCategory_TASK_AGGREGATE {
			for _, model := range models {
				if model.Category != asset.ModelCategory_MODEL_HEAD {
					inputs = append(inputs, model)
				}
			}
		} else {
			inputs = append(inputs, models...)
		}
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

	existingModels, err := s.GetModelDBAL().GetComputeTaskOutputModels(task.Key)
	if err != nil {
		return nil, err
	}
	if err = checkDuplicateModel(existingModels, newModel); err != nil {
		return nil, err
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

	model.Owner = requester

	err = s.GetModelDBAL().AddModel(model)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  model.Key,
		AssetKind: asset.AssetKind_ASSET_MODEL,
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}

	existingModels = append(existingModels, model)
	if areAllOutputsRegistered(task, existingModels) {
		reason := fmt.Sprintf("Last model %s registered by %s", model.Key, requester)
		err = s.GetComputeTaskService().applyTaskAction(task, transitionDone, reason)
		if err != nil {
			return nil, err
		}
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

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_DISABLED,
		AssetKey:  model.Key,
		AssetKind: asset.AssetKind_ASSET_MODEL,
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return err
	}

	return nil
}

type modelCount struct {
	simple uint
	head   uint
	trunk  uint
}

func countModels(models []*asset.Model) modelCount {
	count := modelCount{}

	for _, m := range models {
		switch m.Category {
		case asset.ModelCategory_MODEL_SIMPLE:
			count.simple++
		case asset.ModelCategory_MODEL_HEAD:
			count.head++
		case asset.ModelCategory_MODEL_TRUNK:
			count.trunk++
		}
	}

	return count
}

// areAllOutputsRegistered is based on the cardinality of existingModels to return whether all
// expected outputs are registered or not.
func areAllOutputsRegistered(task *asset.ComputeTask, existingModels []*asset.Model) bool {
	// TOOD: unit test
	count := countModels(existingModels)

	switch task.Category {
	case asset.ComputeTaskCategory_TASK_AGGREGATE, asset.ComputeTaskCategory_TASK_TRAIN:
		return count.simple == 1
	case asset.ComputeTaskCategory_TASK_COMPOSITE:
		return count.head == 1 && count.trunk == 1
	default:
		log.WithField("taskKey", task.Key).WithField("category", task.Category).Warn("unexpected output model check")
		return false
	}
}

// checkDuplicateModel returns an error if a model of the same category already exist
func checkDuplicateModel(existingModels []*asset.Model, model *asset.NewModel) error {
	// TODO: unit test
	for _, m := range existingModels {
		if m.Category == model.Category {
			return fmt.Errorf("%w: task already has a %s model", errors.ErrConflict, model.Category.String())
		}
	}
	return nil
}
