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

	return services.GetModelService().RegisterModel(newModel, mspid)
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