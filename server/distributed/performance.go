package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

// PerformanceAdapter is a grpc server exposing the same Performance interface than standalone mode,
// but relies on a remote chaincode to actually manage the asset.
type PerformanceAdapter struct {
	asset.UnimplementedPerformanceServiceServer
}

// NewPerformanceAdapter creates a Server
func NewPerformanceAdapter() *PerformanceAdapter {
	return &PerformanceAdapter{}
}

func (a *PerformanceAdapter) RegisterPerformance(ctx context.Context, newPerf *asset.NewPerformance) (*asset.Performance, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.performance:RegisterPerformance"

	perf := &asset.Performance{}

	err = invocator.Call(ctx, method, newPerf, perf)

	return perf, err
}

func (a *PerformanceAdapter) GetComputeTaskPerformance(ctx context.Context, param *asset.GetComputeTaskPerformanceParam) (*asset.Performance, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.performance:GetComputeTaskPerformance"

	perf := &asset.Performance{}

	err = invocator.Call(ctx, method, param, perf)

	return perf, err
}
