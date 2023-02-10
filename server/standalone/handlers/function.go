package handlers

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	libCommon "github.com/substra/orchestrator/lib/common"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/interceptors"
)

// FunctionServer is the gRPC facade to Function manipulation
type FunctionServer struct {
	asset.UnimplementedFunctionServiceServer
}

// NewFunctionServer creates a grpc server
func NewFunctionServer() *FunctionServer {
	return &FunctionServer{}
}

// RegisterFunction will persist a new function
func (s *FunctionServer) RegisterFunction(ctx context.Context, a *asset.NewFunction) (*asset.Function, error) {
	log.Ctx(ctx).Debug().Interface("function", a).Msg("Register Function")

	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetFunctionService().RegisterFunction(a, mspid)
}

// GetFunction fetches an function by its key
func (s *FunctionServer) GetFunction(ctx context.Context, params *asset.GetFunctionParam) (*asset.Function, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetFunctionService().GetFunction(params.Key)
}

// QueryFunctions returns a paginated list of all known functions
func (s *FunctionServer) QueryFunctions(ctx context.Context, params *asset.QueryFunctionsParam) (*asset.QueryFunctionsResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	functions, paginationToken, err := services.GetFunctionService().QueryFunctions(libCommon.NewPagination(params.PageToken, params.PageSize), params.Filter)
	if err != nil {
		return nil, err
	}

	return &asset.QueryFunctionsResponse{
		Functions:     functions,
		NextPageToken: paginationToken,
	}, nil
}

// UpdateFunction will update mutable fields of the existing Function. List of mutable fields: name.
func (s *FunctionServer) UpdateFunction(ctx context.Context, params *asset.UpdateFunctionParam) (*asset.UpdateFunctionResponse, error) {
	log.Ctx(ctx).Debug().Interface("function", params).Msg("Update Function")

	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetFunctionService().UpdateFunction(params, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.UpdateFunctionResponse{}, nil
}
