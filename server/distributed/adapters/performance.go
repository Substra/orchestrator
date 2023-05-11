package adapters

import (
	"context"
	"strings"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/server/distributed/interceptors"
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
	invocator, err := interceptors.ExtractInvocator(ctx)
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
		response := &asset.QueryPerformancesResponse{}
		err = invocator.Call(
			ctx,
			"orchestrator.performance:QueryPerformances",
			&asset.QueryPerformancesParam{
				PageToken: "",
				PageSize:  1,
				Filter: &asset.PerformanceQueryFilter{
					ComputeTaskKey:              newPerf.ComputeTaskKey,
					ComputeTaskOutputIdentifier: newPerf.ComputeTaskOutputIdentifier,
				},
			},
			response,
		)
		perf = response.Performances[0]
		return perf, err
	}

	return perf, err
}

func (a *PerformanceAdapter) QueryPerformances(ctx context.Context, param *asset.QueryPerformancesParam) (*asset.QueryPerformancesResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.performance:QueryPerformances"

	response := &asset.QueryPerformancesResponse{}

	err = invocator.Call(ctx, method, param, response)

	return response, err
}
