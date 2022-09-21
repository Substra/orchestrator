package adapters

import (
	"context"
	"strings"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/server/distributed/interceptors"
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
	invocator, err := interceptors.ExtractInvocator(ctx)
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
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.model:GetModel"

	response := &asset.Model{}

	err = invocator.Call(ctx, method, param, response)

	return response, err
}

func (a *ModelAdapter) GetComputeTaskOutputModels(ctx context.Context, param *asset.GetComputeTaskModelsParam) (*asset.GetComputeTaskModelsResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
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

func (a *ModelAdapter) RegisterModels(ctx context.Context, param *asset.RegisterModelsParam) (*asset.RegisterModelsResponse, error) {
	Invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}

	method := "orchestrator.model:RegisterModels"

	response := &asset.RegisterModelsResponse{}

	err = Invocator.Call(ctx, method, param, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
