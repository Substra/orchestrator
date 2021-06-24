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
