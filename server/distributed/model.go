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

package distributed

import (
	"github.com/owkin/orchestrator/lib/asset"
	"golang.org/x/net/context"
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
	method := "org.substra.model:RegisterModel"

	model := &asset.Model{}

	err = invocator.Call(method, newModel, model)

	return model, err
}

func (a *ModelAdapter) GetComputeTaskModels(ctx context.Context, param *asset.GetComputeTaskModelsParam) (*asset.GetComputeTaskModelsResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.model:GetComputeTaskModels"

	response := new(asset.GetComputeTaskModelsResponse)

	err = invocator.Call(method, param, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
