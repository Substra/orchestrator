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
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"
)

// ComputeTaskServer is the gRPC server exposing ComputeTask actions
type ComputeTaskServer struct {
	asset.UnimplementedComputeTaskServiceServer
}

// NewComputeTaskServer creates a Server
func NewComputeTaskServer() *ComputeTaskServer {
	return &ComputeTaskServer{}
}

// RegisterTask will add a new ComputeTask to the network
func (s *ComputeTaskServer) RegisterTask(ctx context.Context, in *asset.NewComputeTask) (*asset.ComputeTask, error) {
	owner, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	task, err := provider.GetComputeTaskService().RegisterTask(in, owner)

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *ComputeTaskServer) RegisterTasks(ctx context.Context, input *asset.RegisterTasksParam) (*asset.RegisterTasksResponse, error) {
	owner, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	tasks, err := provider.GetComputeTaskService().RegisterTasks(input.GetTasks(), owner)
	if err != nil {
		return nil, err
	}

	return &asset.RegisterTasksResponse{
		Tasks: tasks,
	}, nil
}

func (s *ComputeTaskServer) QueryTasks(ctx context.Context, in *asset.QueryTasksParam) (*asset.QueryTasksResponse, error) {
	provider, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	pagination := libCommon.NewPagination(in.PageToken, in.PageSize)

	tasks, paginationToken, err := provider.GetComputeTaskService().QueryTasks(pagination, in.Filter)
	if err != nil {
		return nil, err
	}

	return &asset.QueryTasksResponse{
		Tasks:         tasks,
		NextPageToken: paginationToken,
	}, nil
}

func (s *ComputeTaskServer) GetTask(ctx context.Context, in *asset.GetTaskParam) (*asset.ComputeTask, error) {
	services, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetComputeTaskService().GetTask(in.Key)
}

func (s *ComputeTaskServer) ApplyTaskAction(ctx context.Context, param *asset.ApplyTaskActionParam) (*asset.ApplyTaskActionResponse, error) {
	requester, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = provider.GetComputeTaskService().ApplyTaskAction(param.ComputeTaskKey, param.Action, param.Log, requester)
	if err != nil {
		return nil, err
	}

	return &asset.ApplyTaskActionResponse{}, nil
}
