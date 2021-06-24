package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"

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
	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetPerformanceService().RegisterPerformance(newPerf, mspid)
}

func (s *PerformanceServer) GetComputeTaskPerformance(ctx context.Context, param *asset.GetComputeTaskPerformanceParam) (*asset.Performance, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetPerformanceService().GetComputeTaskPerformance(param.ComputeTaskKey)
}
