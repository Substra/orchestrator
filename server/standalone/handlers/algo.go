package handlers

import (
	"context"

	"github.com/substra/orchestrator/lib/asset"
	libCommon "github.com/substra/orchestrator/lib/common"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/common/logger"
	"github.com/substra/orchestrator/server/standalone/interceptors"
)

// AlgoServer is the gRPC facade to Algo manipulation
type AlgoServer struct {
	asset.UnimplementedAlgoServiceServer
}

// NewAlgoServer creates a grpc server
func NewAlgoServer() *AlgoServer {
	return &AlgoServer{}
}

// RegisterAlgo will persist a new algo
func (s *AlgoServer) RegisterAlgo(ctx context.Context, a *asset.NewAlgo) (*asset.Algo, error) {
	logger.Get(ctx).WithField("algo", a).Debug("Register Algo")

	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetAlgoService().RegisterAlgo(a, mspid)
}

// GetAlgo fetches an algo by its key
func (s *AlgoServer) GetAlgo(ctx context.Context, params *asset.GetAlgoParam) (*asset.Algo, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetAlgoService().GetAlgo(params.Key)
}

// QueryAlgos returns a paginated list of all known algos
func (s *AlgoServer) QueryAlgos(ctx context.Context, params *asset.QueryAlgosParam) (*asset.QueryAlgosResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	algos, paginationToken, err := services.GetAlgoService().QueryAlgos(libCommon.NewPagination(params.PageToken, params.PageSize), params.Filter)
	if err != nil {
		return nil, err
	}

	return &asset.QueryAlgosResponse{
		Algos:         algos,
		NextPageToken: paginationToken,
	}, nil
}

// UpdateAlgo will update mutable fields of the existing Algo. List of mutable fields: name.
func (s *AlgoServer) UpdateAlgo(ctx context.Context, params *asset.UpdateAlgoParam) (*asset.UpdateAlgoResponse, error) {
	logger.Get(ctx).WithField("algo", params).Debug("Update Algo")

	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetAlgoService().UpdateAlgo(params, mspid)
	if err != nil {
		return nil, err
	}

	return &asset.UpdateAlgoResponse{}, nil
}
