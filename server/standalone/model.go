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

package standalone

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
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
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetModelService().RegisterModel(newModel, mspid)
}

func (s *ModelServer) GetComputeTaskOutputModels(ctx context.Context, param *asset.GetComputeTaskModelsParam) (*asset.GetComputeTaskModelsResponse, error) {
	services, err := ExtractProvider(ctx)
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
	services, err := ExtractProvider(ctx)
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
	services, err := ExtractProvider(ctx)
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
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetModelService().DisableModel(param.ModelKey, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.DisableModelResponse{}, nil
}
