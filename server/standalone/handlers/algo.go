package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// AlgoServer is the gRPC facade to Algo manipulation
type AlgoServer struct {
	asset.UnimplementedAlgoServiceServer
}

// NewAlgoServer creates a grpc server
func NewAlgoServer() *AlgoServer {
	return &AlgoServer{}
}

// RegisterAlgo will persiste a new algo
func (s *AlgoServer) RegisterAlgo(ctx context.Context, a *asset.NewAlgo) (*asset.Algo, error) {
	logger.Get(ctx).WithField("algo", a).Debug("Register Algo")

	mspid, err := common.ExtractMSPID(ctx)
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

	algos, paginationToken, err := services.GetAlgoService().QueryAlgos(params.Category, libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryAlgosResponse{
		Algos:         algos,
		NextPageToken: paginationToken,
	}, nil
}
