package service

import (
	"fmt"
	"sort"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ModelAPI interface {
	GetComputeTaskOutputModels(key string) ([]*asset.Model, error)
	GetComputeTaskInputModels(key string) ([]*asset.Model, error)
	CanDisableModel(key, requester string) (bool, error)
	// DisableModel removes a model address and emit a "disabled" event
	DisableModel(key string, requester string) error
	GetModel(key string) (*asset.Model, error)
	QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error)
	RegisterModels(models []*asset.NewModel, owner string) ([]*asset.Model, error)
}

type ModelServiceProvider interface {
	GetModelService() ModelAPI
}

type ModelDependencyProvider interface {
	LoggerProvider
	persistence.ModelDBALProvider
	persistence.AlgoDBALProvider
	persistence.DataManagerDBALProvider
	PermissionServiceProvider
	ComputeTaskServiceProvider
	ComputePlanServiceProvider
	EventServiceProvider
	TimeServiceProvider
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
	s.GetLogger().WithField("key", key).Debug("Get model")
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

		switch task.Category {
		case asset.ComputeTaskCategory_TASK_AGGREGATE:
			for _, model := range models {
				if model.Category == asset.ModelCategory_MODEL_SIMPLE {
					inputs = append(inputs, model)
				}
			}
		case asset.ComputeTaskCategory_TASK_COMPOSITE:
			// For this function the order of assets is important we should always have the HEAD MODEL first in the list
			// Otherwise we end up feeding the head and trunk from the previous composite, ignoring the aggregate
			sort.SliceStable(models, func(i, j int) bool {
				return models[i].Category == asset.ModelCategory_MODEL_HEAD
			})
			// true if the parent has contributed an input to the composite task
			parentContributed := false
			for _, model := range models {
				// Head model should always come from the first parent possible
				if model.Category == asset.ModelCategory_MODEL_HEAD && !containsHeadModel(inputs) {
					inputs = append(inputs, model)
					parentContributed = true
				}

				singleParent := len(task.ParentTaskKeys) == 1
				completeInputs := len(inputs) < 2
				// Add trunk from parent if it's a single parent or if we still miss an input and the parent has not contributed a model yet
				// Current parent should contribute the trunk model if:
				// - it's a single parent
				// - it has not contributed yet but not all inputs are set
				shouldContributeTrunk := singleParent || (!parentContributed && completeInputs)

				if model.Category == asset.ModelCategory_MODEL_SIMPLE && shouldContributeTrunk {
					inputs = append(inputs, model)
					parentContributed = true
				}
			}
		default:
			inputs = append(inputs, models...)
		}
	}

	return inputs, nil
}

func (s *ModelService) registerModel(newModel *asset.NewModel, requester string, existingModels []*asset.Model, task *asset.ComputeTask) (*asset.Model, error) {
	s.GetLogger().WithField("model", newModel).WithField("requester", requester).Debug("Registering new model")

	err := newModel.Validate()
	if err != nil {
		return nil, errors.FromValidationError(asset.ModelKind, err)
	}

	if task.Worker != requester {
		return nil, errors.NewPermissionDenied(fmt.Sprintf("only %q worker can register model", task.Worker))
	}

	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot register model for task with status %q", task.Status.String()))
	}

	if err = checkDuplicateModel(existingModels, newModel); err != nil {
		return nil, err
	}

	var model *asset.Model

	switch task.Category {
	case asset.ComputeTaskCategory_TASK_TRAIN, asset.ComputeTaskCategory_TASK_AGGREGATE:
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
		return nil, errors.NewUnimplemented("unhandled model category")
	}

	model.Owner = requester
	model.CreationDate = timestamppb.New(s.GetTimeService().GetTransactionTime())

	err = s.GetModelDBAL().AddModel(model)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  model.Key,
		AssetKind: asset.AssetKind_ASSET_MODEL,
		Asset:     &asset.Event_Model{Model: model},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (s *ModelService) registerSimpleModel(newModel *asset.NewModel, requester string, task *asset.ComputeTask) (*asset.Model, error) {
	// This should be checked by caller, but better safe than sorry
	if !(task.Category == asset.ComputeTaskCategory_TASK_TRAIN || task.Category == asset.ComputeTaskCategory_TASK_AGGREGATE) {
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot register train model for %q task", task.Category.String()))
	}
	if newModel.Category != asset.ModelCategory_MODEL_SIMPLE {
		return nil, errors.NewBadRequest("cannot register non-simple model")
	}

	var permissions *asset.Permissions

	switch task.Category {
	case asset.ComputeTaskCategory_TASK_TRAIN:
		permissions = task.GetTrain().ModelPermissions
	case asset.ComputeTaskCategory_TASK_AGGREGATE:
		permissions = task.GetAggregate().ModelPermissions
	default:
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot set model permissions for %q task", task.Category.String()))
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
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot register composite model for %q task", task.Category.String()))
	}
	if !(newModel.Category == asset.ModelCategory_MODEL_HEAD || newModel.Category == asset.ModelCategory_MODEL_SIMPLE) {
		return nil, errors.NewBadRequest("cannot register non-composite model")
	}

	var permissions *asset.Permissions

	switch newModel.Category {
	case asset.ModelCategory_MODEL_HEAD:
		permissions = task.GetComposite().HeadPermissions
	case asset.ModelCategory_MODEL_SIMPLE:
		permissions = task.GetComposite().TrunkPermissions
	default:
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot set permissions for %q model", newModel.Category.String()))
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
	s.GetLogger().WithField("model", key).Debug("checking whether model can be disabled")
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
	s.GetLogger().WithField("modelKey", key).Debug("disabling model")
	model, err := s.GetModelDBAL().GetModel(key)
	if err != nil {
		return err
	}

	canClean, err := s.canDisableModel(model, requester)
	if err != nil {
		return err
	}
	if !canClean {
		return errors.NewCannotDisableModel("cannot disable a model in use")
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
		Asset:     &asset.Event_Model{Model: model},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return err
	}

	return nil
}

// AreAllOutputsRegistered is based on the cardinality of existingModels to return whether all
// expected outputs are registered or not.
func (s *ModelService) AreAllOutputsRegistered(task *asset.ComputeTask, existingModels []*asset.Model) bool {
	count := countModels(existingModels)

	switch task.Category {
	case asset.ComputeTaskCategory_TASK_TRAIN:
		return count.simple == 1
	case asset.ComputeTaskCategory_TASK_COMPOSITE:
		return count.head == 1 && count.simple == 1
	case asset.ComputeTaskCategory_TASK_AGGREGATE:
		return count.simple == 1
	default:
		s.GetLogger().WithField("taskKey", task.Key).WithField("category", task.Category).Warn("unexpected output model check")
		return false
	}
}

func (s *ModelService) RegisterModels(models []*asset.NewModel, owner string) ([]*asset.Model, error) {
	s.GetLogger().WithField("owner", owner).WithField("nbModels", len(models)).Debug("Registering models")

	registeredModels := make([]*asset.Model, len(models))

	existingModels, err := s.GetModelDBAL().GetComputeTaskOutputModels(models[0].ComputeTaskKey)
	if err != nil {
		return nil, err
	}

	for modelIndex, newModel := range models {
		computeTask, err := s.GetComputeTaskService().GetTask(models[0].ComputeTaskKey)
		if err != nil {
			return nil, err
		}

		if newModel.ComputeTaskKey != computeTask.Key {
			return nil, errors.NewBadRequest("All models should be part of the same task")
		}

		model, err := s.registerModel(newModel, owner, existingModels, computeTask)
		if err != nil {
			return nil, err
		}
		registeredModels[modelIndex] = model

		existingModels = append(existingModels, model)
		if s.AreAllOutputsRegistered(computeTask, existingModels) {
			reason := fmt.Sprintf("Last model %s registered by %s", model.Key, owner)
			err = s.GetComputeTaskService().applyTaskAction(computeTask, transitionDone, reason)
			if err != nil {
				return nil, err
			}
		}
	}

	return registeredModels, nil
}

type modelCount struct {
	simple uint
	head   uint
}

func countModels(models []*asset.Model) modelCount {
	count := modelCount{}

	for _, m := range models {
		switch m.Category {
		case asset.ModelCategory_MODEL_SIMPLE:
			count.simple++
		case asset.ModelCategory_MODEL_HEAD:
			count.head++
		}
	}

	return count
}

// checkDuplicateModel returns an error if a model of the same category already exist
func checkDuplicateModel(existingModels []*asset.Model, model *asset.NewModel) error {
	for _, m := range existingModels {
		if m.Category == model.Category {
			return errors.NewError(errors.ErrConflict, fmt.Sprintf("task already has a %s model", model.Category.String()))
		}
	}
	return nil
}

// containsHeadModel returns true if the slice contains a HEAD model
func containsHeadModel(inputs []*asset.Model) bool {
	for _, m := range inputs {
		if m.Category == asset.ModelCategory_MODEL_HEAD {
			return true
		}
	}

	return false
}
