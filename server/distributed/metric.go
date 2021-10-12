package distributed

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
)

// MetricAdapter is a grpc server exposing the same node interface,
// but relies on a remote chaincode to actually manage the asset.
type MetricAdapter struct {
	asset.UnimplementedMetricServiceServer
}

// NewMetricAdapter creates a Server
func NewMetricAdapter() *MetricAdapter {
	return &MetricAdapter{}
}

// RegisterMetric will add a new Metric to the network
func (a *MetricAdapter) RegisterMetric(ctx context.Context, in *asset.NewMetric) (*asset.Metric, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.metric:RegisterMetric"

	response := &asset.Metric{}

	err = invocator.Call(ctx, method, in, response)

	if err != nil && isFabricTimeoutRetry(ctx) && strings.Contains(err.Error(), errors.ErrConflict) {
		// In this very specific case we are in a retry context after a timeout.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		err = invocator.Call(ctx, "orchestrator.metric:GetMetric", &asset.GetMetricParam{Key: in.Key}, response)
		return response, err
	}

	return response, err
}

// GetMetric returns an metric from its key
func (a *MetricAdapter) GetMetric(ctx context.Context, query *asset.GetMetricParam) (*asset.Metric, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.metric:GetMetric"

	response := &asset.Metric{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}

// QueryMetrics returns all known metrics
func (a *MetricAdapter) QueryMetrics(ctx context.Context, query *asset.QueryMetricsParam) (*asset.QueryMetricsResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.metric:QueryMetrics"

	response := &asset.QueryMetricsResponse{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}
