package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// MetricServer is the gRPC facade to Metric manipulation
type MetricServer struct {
	asset.UnimplementedMetricServiceServer
}

// NewMetricServer creates a grpc server
func NewMetricServer() *MetricServer {
	return &MetricServer{}
}

// RegisterMetric will persiste a new metric
func (s *MetricServer) RegisterMetric(ctx context.Context, o *asset.NewMetric) (*asset.Metric, error) {
	logger.Get(ctx).WithField("metric", o).Debug("register metric")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetMetricService().RegisterMetric(o, mspid)
}

// GetMetric fetches an metric by its key
func (s *MetricServer) GetMetric(ctx context.Context, params *asset.GetMetricParam) (*asset.Metric, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetMetricService().GetMetric(params.Key)
}

// QueryMetrics returns a paginated list of all known metrics
func (s *MetricServer) QueryMetrics(ctx context.Context, params *asset.QueryMetricsParam) (*asset.QueryMetricsResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	metrics, paginationToken, err := services.GetMetricService().QueryMetrics(libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryMetricsResponse{
		Metrics:       metrics,
		NextPageToken: paginationToken,
	}, nil
}
