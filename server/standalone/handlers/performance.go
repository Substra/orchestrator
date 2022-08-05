package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	commonInterceptors "github.com/owkin/orchestrator/server/common/interceptors"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// PerformanceServer is the gRPC facade to Performance manipulation
type PerformanceServer struct {
	asset.UnimplementedPerformanceServiceServer
}

// NewPerformanceServer creates a grpc server
func NewPerformanceServer() *PerformanceServer {
	return &PerformanceServer{}
}

func (s *PerformanceServer) RegisterPerformance(ctx context.Context, newPerf *asset.NewPerformance) (*asset.Performance, error) {
	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetPerformanceService().RegisterPerformance(newPerf, mspid)
}

func (s *PerformanceServer) QueryPerformances(ctx context.Context, param *asset.QueryPerformancesParam) (*asset.QueryPerformancesResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	performances, paginationToken, err := services.GetPerformanceService().QueryPerformances(libCommon.NewPagination(param.PageToken, param.PageSize), param.Filter)
	if err != nil {
		return nil, err
	}

	return &asset.QueryPerformancesResponse{
		Performances:  performances,
		NextPageToken: paginationToken,
	}, nil
}
