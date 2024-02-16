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
	FunctionServiceProvider
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

	if task.Status != asset.ComputeTaskStatus_STATUS_EXECUTING {
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot register model for task with status %q", task.Status.String()))
	}

	taskOutput, ok := task.Outputs[newModel.ComputeTaskOutputIdentifier]
	if !ok {
		return nil, errors.NewMissingTaskOutput(task.Key, newModel.ComputeTaskOutputIdentifier)
	}
	function, err := s.GetFunctionService().GetFunction(task.FunctionKey)
	if err != nil {
		return nil, err
	}
	functionOutput, ok := function.Outputs[newModel.ComputeTaskOutputIdentifier]
	if !ok {
		// This should never happen since task outputs are checked against function on registration
		return nil, errors.NewInternal(fmt.Sprintf("missing function output %q for task %q", newModel.ComputeTaskOutputIdentifier, task.Key))
	}
	if functionOutput.Kind != asset.AssetKind_ASSET_MODEL {
		return nil, errors.NewIncompatibleTaskOutput(task.Key, newModel.ComputeTaskOutputIdentifier, functionOutput.Kind.String(), asset.AssetKind_ASSET_MODEL.String())
	}

	if outputCounter[newModel.ComputeTaskOutputIdentifier] >= 1 && !functionOutput.Multiple {
		return nil, errors.NewError(orcerrors.ErrConflict, fmt.Sprintf("compute task %q already has its unique output %q registered", task.Key, newModel.ComputeTaskOutputIdentifier))
	}

	model := &asset.Model{
		Key:            newModel.Key,
		ComputeTaskKey: task.Key,
		Address:        newModel.Address,
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
