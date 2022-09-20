package handlers

import (
	"context"

	"github.com/substra/orchestrator/lib/asset"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/interceptors"
)

// ModelServer is the gRPC facade to Model manipulation
type ModelServer struct {
	asset.UnimplementedModelServiceServer
}

// NewModelServer creates a grpc server
func NewModelServer() *ModelServer {
	return &ModelServer{}
}

func (s *ModelServer) RegisterModel(ctx context.Context, newModel *asset.NewModel) (*asset.Model, error) {
	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	models, err := services.GetModelService().RegisterModels([]*asset.NewModel{newModel}, mspid)
	if err != nil {
		return nil, err
	}

	return models[0], err
}

func (s *ModelServer) GetModel(ctx context.Context, in *asset.GetModelParam) (*asset.Model, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetModelService().GetModel(in.Key)
}

func (s *ModelServer) GetComputeTaskOutputModels(ctx context.Context, param *asset.GetComputeTaskModelsParam) (*asset.GetComputeTaskModelsResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	models, err := services.GetModelService().GetComputeTaskOutputModels(param.ComputeTaskKey)
	if err != nil {
		return nil, err
	}

	return &asset.GetComputeTaskModelsResponse{
		Models: models,
	}, nil
}

func (s *ModelServer) CanDisableModel(ctx context.Context, param *asset.CanDisableModelParam) (*asset.CanDisableModelResponse, error) {
	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	can, err := services.GetModelService().CanDisableModel(param.ModelKey, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.CanDisableModelResponse{
		CanDisable: can,
	}, nil
}

func (s *ModelServer) RegisterModels(ctx context.Context, param *asset.RegisterModelsParam) (*asset.RegisterModelsResponse, error) {
	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	models, err := services.GetModelService().RegisterModels(param.Models, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.RegisterModelsResponse{
		Models: models,
	}, nil
}
