package handlers

import (
	"context"

	"github.com/substra/orchestrator/lib/asset"
	libCommon "github.com/substra/orchestrator/lib/common"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/interceptors"
)

// ComputeTaskServer is the gRPC server exposing ComputeTask actions
type ComputeTaskServer struct {
	asset.UnimplementedComputeTaskServiceServer
}

// NewComputeTaskServer creates a Server
func NewComputeTaskServer() *ComputeTaskServer {
	return &ComputeTaskServer{}
}

func (s *ComputeTaskServer) RegisterTasks(ctx context.Context, input *asset.RegisterTasksParam) (*asset.RegisterTasksResponse, error) {
	owner, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	tasks, err := provider.GetComputeTaskService().RegisterTasks(input.GetTasks(), owner)

	if err != nil {
		return nil, err
	}

	return &asset.RegisterTasksResponse{Tasks: tasks}, nil
}

func (s *ComputeTaskServer) QueryTasks(ctx context.Context, in *asset.QueryTasksParam) (*asset.QueryTasksResponse, error) {
	provider, err := interceptors.ExtractProvider(ctx)
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
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetComputeTaskService().GetTask(in.Key)
}

func (s *ComputeTaskServer) ApplyTaskAction(ctx context.Context, param *asset.ApplyTaskActionParam) (*asset.ApplyTaskActionResponse, error) {
	requester, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = provider.GetComputeTaskService().ApplyTaskAction(param.ComputeTaskKey, param.Action, param.Log, requester)
	if err != nil {
		return nil, err
	}

	return &asset.ApplyTaskActionResponse{}, nil
}

func (s *ComputeTaskServer) GetTaskInputAssets(ctx context.Context, param *asset.GetTaskInputAssetsParam) (*asset.GetTaskInputAssetsResponse, error) {
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	inputs, err := provider.GetComputeTaskService().GetInputAssets(param.ComputeTaskKey)
	if err != nil {
		return nil, err
	}

	return &asset.GetTaskInputAssetsResponse{Assets: inputs}, nil
}

func (s *ComputeTaskServer) DisableOutput(ctx context.Context, param *asset.DisableOutputParam) (*asset.DisableOutputResponse, error) {
	requester, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = provider.GetComputeTaskService().DisableOutput(param.ComputeTaskKey, param.Identifier, requester)
	if err != nil {
		return nil, err
	}

	return &asset.DisableOutputResponse{}, nil
}
