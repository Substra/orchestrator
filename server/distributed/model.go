package distributed

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
)

// ModelAdapter is a grpc server exposing the same Model interface,
// but relies on a remote chaincode to actually manage the asset.
type ModelAdapter struct {
	asset.UnimplementedModelServiceServer
}

// NewModelAdapter creates a Server
func NewModelAdapter() *ModelAdapter {
	return &ModelAdapter{}
}

func (a *ModelAdapter) RegisterModel(ctx context.Context, newModel *asset.NewModel) (*asset.Model, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.model:RegisterModel"

	model := &asset.Model{}

	err = invocator.Call(ctx, method, newModel, model)

	if err != nil && isFabricTimeoutRetry(ctx) && strings.Contains(err.Error(), errors.ErrConflict) {
		// In this very specific case we are in a retry context after a timeout.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		err = invocator.Call(ctx, "orchestrator.model:GetModel", &asset.GetModelParam{Key: newModel.Key}, model)
		return model, err
	}

	return model, err
}

func (a *ModelAdapter) GetModel(ctx context.Context, param *asset.GetModelParam) (*asset.Model, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.model:GetModel"

	response := &asset.Model{}

	err = invocator.Call(ctx, method, param, response)

	return response, err
}

// QueryModels returns all known models
func (a *ModelAdapter) QueryModels(ctx context.Context, query *asset.QueryModelsParam) (*asset.QueryModelsResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.model:QueryModels"

	response := &asset.QueryModelsResponse{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}

func (a *ModelAdapter) GetComputeTaskOutputModels(ctx context.Context, param *asset.GetComputeTaskModelsParam) (*asset.GetComputeTaskModelsResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.model:GetComputeTaskOutputModels"

	response := new(asset.GetComputeTaskModelsResponse)

	err = invocator.Call(ctx, method, param, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (a *ModelAdapter) GetComputeTaskInputModels(ctx context.Context, param *asset.GetComputeTaskModelsParam) (*asset.GetComputeTaskModelsResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.model:GetComputeTaskInputModels"

	response := new(asset.GetComputeTaskModelsResponse)

	err = invocator.Call(ctx, method, param, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (a *ModelAdapter) CanDisableModel(ctx context.Context, param *asset.CanDisableModelParam) (*asset.CanDisableModelResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.model:CanDisableModel"

	response := new(asset.CanDisableModelResponse)

	err = invocator.Call(ctx, method, param, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (a *ModelAdapter) DisableModel(ctx context.Context, param *asset.DisableModelParam) (*asset.DisableModelResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.model:DisableModel"

	err = invocator.Call(ctx, method, param, nil)
	if err != nil {
		return nil, err
	}

	return &asset.DisableModelResponse{}, nil
}
