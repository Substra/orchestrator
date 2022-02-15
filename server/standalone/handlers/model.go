package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
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
	mspid, err := common.ExtractMSPID(ctx)
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

// QueryModels returns a paginated list of all known models
func (s *ModelServer) QueryModels(ctx context.Context, params *asset.QueryModelsParam) (*asset.QueryModelsResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	models, paginationToken, err := services.GetModelService().QueryModels(params.Category, libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryModelsResponse{
		Models:        models,
		NextPageToken: paginationToken,
	}, nil
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

func (s *ModelServer) GetComputeTaskInputModels(ctx context.Context, param *asset.GetComputeTaskModelsParam) (*asset.GetComputeTaskModelsResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	models, err := services.GetModelService().GetComputeTaskInputModels(param.ComputeTaskKey)
	if err != nil {
		return nil, err
	}

	return &asset.GetComputeTaskModelsResponse{
		Models: models,
	}, nil
}

func (s *ModelServer) CanDisableModel(ctx context.Context, param *asset.CanDisableModelParam) (*asset.CanDisableModelResponse, error) {
	mspid, err := common.ExtractMSPID(ctx)
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

func (s *ModelServer) DisableModel(ctx context.Context, param *asset.DisableModelParam) (*asset.DisableModelResponse, error) {
	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetModelService().DisableModel(param.ModelKey, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.DisableModelResponse{}, nil
}

func (s *ModelServer) RegisterModels(ctx context.Context, param *asset.RegisterModelsParam) (*asset.RegisterModelsResponse, error) {
	mspid, err := common.ExtractMSPID(ctx)
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
