package distributed

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
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

	if err != nil && isFabricTimeoutRetry(ctx) && strings.Contains(err.Error(), errors.ErrConflict) {
		// In this very specific case we are in a retry context after a timeout.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		err = invocator.Call(
			ctx,
			"orchestrator.performance:GetComputeTaskPerformance",
			&asset.GetComputeTaskPerformanceParam{ComputeTaskKey: newPerf.ComputeTaskKey},
			perf,
		)
		return perf, err
	}

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
