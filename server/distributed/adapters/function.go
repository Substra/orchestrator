package adapters

import (
	"context"
	"strings"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/server/distributed/interceptors"
)

// FunctionAdapter is a grpc server exposing the same function interface than standalone,
// but relies on a remote chaincode to actually manage the asset.
type FunctionAdapter struct {
	asset.UnimplementedFunctionServiceServer
}

// NewFunctionAdapter creates a Server
func NewFunctionAdapter() *FunctionAdapter {
	return &FunctionAdapter{}
}

// RegisterFunction will add a new Function to the network
func (a *FunctionAdapter) RegisterFunction(ctx context.Context, in *asset.NewFunction) (*asset.Function, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.function:RegisterFunction"

	response := &asset.Function{}

	err = invocator.Call(ctx, method, in, response)

	if err != nil && isFabricTimeoutRetry(ctx) && strings.Contains(err.Error(), errors.ErrConflict) {
		// In this very specific case we are in a retry context after a timeout.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		err = invocator.Call(ctx, "orchestrator.function:GetFunction", &asset.GetFunctionParam{Key: in.Key}, response)
		return response, err
	}

	return response, err
}

// GetFunction returns an function from its key
func (a *FunctionAdapter) GetFunction(ctx context.Context, query *asset.GetFunctionParam) (*asset.Function, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.function:GetFunction"

	response := &asset.Function{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}

// QueryFunctions returns all known functions
func (a *FunctionAdapter) QueryFunctions(ctx context.Context, query *asset.QueryFunctionsParam) (*asset.QueryFunctionsResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.function:QueryFunctions"

	response := &asset.QueryFunctionsResponse{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}

// UpdateFunction will update an Function
func (a *FunctionAdapter) UpdateFunction(ctx context.Context, query *asset.UpdateFunctionParam) (*asset.UpdateFunctionResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.function:UpdateFunction"

	response := &asset.UpdateFunctionResponse{}

	err = invocator.Call(ctx, method, query, nil)

	return response, err
}

func (a *FunctionAdapter) ApplyFunctionAction(ctx context.Context, param *asset.ApplyFunctionActionParam) (*asset.ApplyFunctionActionResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computetask:ApplyFunctionAction"

	err = invocator.Call(ctx, method, param, nil)
	if err != nil {
		return nil, err
	}

	return &asset.ApplyFunctionActionResponse{}, nil
}
