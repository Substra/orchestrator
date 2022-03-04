package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/utils"
	"github.com/stretchr/testify/assert"
)

func TestMetricAdapterImplementServer(t *testing.T) {
	adapter := NewMetricAdapter()
	assert.Implementsf(t, (*asset.MetricServiceServer)(nil), adapter, "MetricAdapter should implements MetricServiceServer")
}

func TestRegisterMetric(t *testing.T) {
	adapter := NewMetricAdapter()

	newObj := &asset.NewMetric{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.metric:RegisterMetric", newObj, &asset.Metric{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterMetric(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestGetMetric(t *testing.T) {
	adapter := NewMetricAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.GetMetricParam{Key: "uuid"}

	invocator.On("Call", utils.AnyContext, "orchestrator.metric:GetMetric", param, &asset.Metric{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.GetMetric(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryMetrics(t *testing.T) {
	adapter := NewMetricAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.QueryMetricsParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", utils.AnyContext, "orchestrator.metric:QueryMetrics", param, &asset.QueryMetricsResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryMetrics(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
