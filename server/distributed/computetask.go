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
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

// ComputeTaskAdapter is a grpc server exposing the same task interface,
// but relies on a remote chaincode to actually manage the asset.
type ComputeTaskAdapter struct {
	asset.UnimplementedComputeTaskServiceServer
}

// NewComputeTaskAdapter creates a Server
func NewComputeTaskAdapter() *ComputeTaskAdapter {
	return &ComputeTaskAdapter{}
}

func (a *ComputeTaskAdapter) RegisterTask(ctx context.Context, in *asset.NewComputeTask) (*asset.ComputeTask, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.computetask:RegisterTask"

	response := &asset.ComputeTask{}

	err = invocator.Call(method, in, response)

	return response, err
}

func (a *ComputeTaskAdapter) QueryTasks(ctx context.Context, param *asset.QueryTasksParam) (*asset.QueryTasksResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "org.substra.computetask:QueryTasks"

	response := &asset.QueryTasksResponse{}

	err = invocator.Call(method, param, response)

	return response, err
}
