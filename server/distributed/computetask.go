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

// RegisterTask performs validation and add a new task to a compute plan
func (a *ComputeTaskAdapter) RegisterTask(ctx context.Context, in *asset.NewComputeTask) (*asset.ComputeTask, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computetask:RegisterTask"

	response := &asset.ComputeTask{}

	err = invocator.Call(method, in, response)

	return response, err
}

// RegisterTasks processes a batch of new tasks to add them to a compute plan
func (a *ComputeTaskAdapter) RegisterTasks(ctx context.Context, input *asset.RegisterTasksParam) (*asset.RegisterTasksResponse, error) {
	Invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computetask:RegisterTasks"

	response := &asset.RegisterTasksResponse{}

	err = Invocator.Call(method, input, nil)

	return response, err
}

// GetTask returns a task from its key
func (a *ComputeTaskAdapter) GetTask(ctx context.Context, query *asset.GetTaskParam) (*asset.ComputeTask, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computetask:GetTask"

	response := &asset.ComputeTask{}

	err = invocator.Call(method, query, response)

	return response, err
}

// QueryTasks returns tasks matching the selection criteria
func (a *ComputeTaskAdapter) QueryTasks(ctx context.Context, param *asset.QueryTasksParam) (*asset.QueryTasksResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computetask:QueryTasks"

	response := &asset.QueryTasksResponse{}

	err = invocator.Call(method, param, response)

	return response, err
}

// ApplyTaskAction updates a task status
func (a *ComputeTaskAdapter) ApplyTaskAction(ctx context.Context, param *asset.ApplyTaskActionParam) (*asset.ApplyTaskActionResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computetask:ApplyTaskAction"

	err = invocator.Call(method, param, nil)
	if err != nil {
		return nil, err
	}

	return &asset.ApplyTaskActionResponse{}, nil
}
