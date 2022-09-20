package service

import (
	"fmt"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ModelAPI interface {
	GetComputeTaskOutputModels(key string) ([]*asset.Model, error)
	CanDisableModel(key, requester string) (bool, error)
	GetModel(key string) (*asset.Model, error)
	RegisterModels(models []*asset.NewModel, owner string) ([]*asset.Model, error)
	GetCheckedModel(key string, worker string) (*asset.Model, error)
	disable(assetKey string) error
}

type ModelServiceProvider interface {
	GetModelService() ModelAPI
}

type ModelDependencyProvider interface {
	LoggerProvider
	persistence.ModelDBALProvider
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
	s.GetLogger().Debug().Str("key", key).Msg("Get model")
	return s.GetModelDBAL().GetModel(key)
}

// GetCheckedModel returns the model if it exists and it can be processed by the worker
func (s *ModelService) GetCheckedModel(key string, worker string) (*asset.Model, error) {
	s.GetLogger().Debug().Str("key", key).Msg("Get model")
	model, err := s.GetModelDBAL().GetModel(key)
	if err != nil {
		return nil, err
	}
	if ok := s.GetPermissionService().CanProcess(model.Permissions, worker); !ok {
		return nil, orcerrors.NewPermissionDenied(fmt.Sprintf("not authorized to process model %q", model.Key))
	}
	return model, nil
}

func (s *ModelService) registerModel(newModel *asset.NewModel, requester string, outputCounter persistence.ComputeTaskOutputCounter, task *asset.ComputeTask) (*asset.Model, error) {
	s.GetLogger().Debug().Interface("model", newModel).Str("requester", requester).Msg("Registering new model")

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

	taskOutput, ok := task.Outputs[newModel.ComputeTaskOutputIdentifier]
	if !ok {
		return nil, errors.NewMissingTaskOutput(task.Key, newModel.ComputeTaskOutputIdentifier)
	}
	algoOutput, ok := task.Algo.Outputs[newModel.ComputeTaskOutputIdentifier]
	if !ok {
		// This should never happen since task outputs are checked against algo on registration
		return nil, errors.NewInternal(fmt.Sprintf("missing algo output %q for task %q", newModel.ComputeTaskOutputIdentifier, task.Key))
	}
	if algoOutput.Kind != asset.AssetKind_ASSET_MODEL {
		return nil, errors.NewIncompatibleTaskOutput(task.Key, newModel.ComputeTaskOutputIdentifier, algoOutput.Kind.String(), asset.AssetKind_ASSET_MODEL.String())
	}

	if outputCounter[newModel.ComputeTaskOutputIdentifier] >= 1 && !algoOutput.Multiple {
		return nil, errors.NewError(orcerrors.ErrConflict, fmt.Sprintf("compute task %q already has its unique output %q registered", task.Key, newModel.ComputeTaskOutputIdentifier))
	}

	var model *asset.Model

	switch task.Category {
	case asset.ComputeTaskCategory_TASK_TRAIN, asset.ComputeTaskCategory_TASK_AGGREGATE, asset.ComputeTaskCategory_TASK_PREDICT:
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

	model.Permissions = taskOutput.Permissions
	model.Owner = requester
	model.CreationDate = timestamppb.New(s.GetTimeService().GetTransactionTime())

	err = s.GetModelDBAL().AddModel(model, newModel.ComputeTaskOutputIdentifier)
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
	outputAsset := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              newModel.ComputeTaskKey,
		ComputeTaskOutputIdentifier: newModel.ComputeTaskOutputIdentifier,
		AssetKind:                   asset.AssetKind_ASSET_MODEL,
		AssetKey:                    newModel.Key,
	}
	err = s.GetComputeTaskService().addComputeTaskOutputAsset(outputAsset)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func (s *ModelService) registerSimpleModel(newModel *asset.NewModel, requester string, task *asset.ComputeTask) (*asset.Model, error) {
	// This should be checked by caller, but better safe than sorry
	if !(task.Category == asset.ComputeTaskCategory_TASK_TRAIN || task.Category == asset.ComputeTaskCategory_TASK_AGGREGATE || task.Category == asset.ComputeTaskCategory_TASK_PREDICT) {
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot register train model for %q task", task.Category.String()))
	}
	if newModel.Category != asset.ModelCategory_MODEL_SIMPLE {
		return nil, errors.NewBadRequest("cannot register non-simple model")
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
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot register composite model for %q task", task.Category.String()))
	}
	if !(newModel.Category == asset.ModelCategory_MODEL_HEAD || newModel.Category == asset.ModelCategory_MODEL_SIMPLE) {
		return nil, errors.NewBadRequest("cannot register non-composite model")
	}

	model := &asset.Model{
		Key:            newModel.Key,
		Category:       newModel.Category,
		ComputeTaskKey: task.Key,
		Address:        newModel.Address,
	}

	return model, nil
}

// CanDisableModel returns true if the model can be disabled
func (s *ModelService) CanDisableModel(key string, requester string) (bool, error) {
	s.GetLogger().Debug().Str("modelKey", key).Msg("checking whether model can be disabled")
	model, err := s.GetModelDBAL().GetModel(key)
	if err != nil {
		return false, err
	}

	return s.canDisableModel(model, requester)
}

func (s *ModelService) canDisableModel(model *asset.Model, requester string) (bool, error) {
	return s.GetComputeTaskService().canDisableModels(model.ComputeTaskKey, requester)
}

func (s *ModelService) disable(assetKey string) error {
	s.GetLogger().Debug().Str("modelKey", assetKey).Msg("disabling model")
	model, err := s.GetModelDBAL().GetModel(assetKey)
	if err != nil {
		return err
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
	return s.GetEventService().RegisterEvents(event)
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
	case asset.ComputeTaskCategory_TASK_PREDICT:
		return count.simple == 1
	default:
		s.GetLogger().Warn().Str("taskKey", task.Key).Str("category", task.Category.String()).Msg("unexpected output model check")
		return false
	}
}

func (s *ModelService) RegisterModels(models []*asset.NewModel, owner string) ([]*asset.Model, error) {
	s.GetLogger().Debug().Str("owner", owner).Int("nbModels", len(models)).Msg("Registering models")

	registeredModels := make([]*asset.Model, len(models))

	outputCounter, err := s.GetComputeTaskService().getTaskOutputCounter(models[0].ComputeTaskKey)
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

		model, err := s.registerModel(newModel, owner, outputCounter, computeTask)
		if err != nil {
			return nil, err
		}
		registeredModels[modelIndex] = model

		outputCounter[newModel.ComputeTaskOutputIdentifier]++
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
