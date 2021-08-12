package distributed

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
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

	err = Invocator.Call(ctx, method, input, nil)

	if err != nil && isFabricTimeoutRetry(ctx) && len(input.Tasks) == 1 && strings.Contains(err.Error(), errors.ErrConflict.Error()) {
		// In this very specific case we are in a retry context after a timeout and the batch only contains a single task.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		return response, nil
	}

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

	err = invocator.Call(ctx, method, query, response)

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

	err = invocator.Call(ctx, method, param, response)

	return response, err
}

// ApplyTaskAction updates a task status
func (a *ComputeTaskAdapter) ApplyTaskAction(ctx context.Context, param *asset.ApplyTaskActionParam) (*asset.ApplyTaskActionResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computetask:ApplyTaskAction"

	err = invocator.Call(ctx, method, param, nil)
	if err != nil {
		return nil, err
	}

	return &asset.ApplyTaskActionResponse{}, nil
}
