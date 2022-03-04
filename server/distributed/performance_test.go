package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPerformanceAdapterImplementServer(t *testing.T) {
	adapter := NewPerformanceAdapter()
	assert.Implements(t, (*asset.PerformanceServiceServer)(nil), adapter)
}

func TestRegisterPerformance(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.NewPerformance{}

	invocator.On("Call", utils.AnyContext, "orchestrator.performance:RegisterPerformance", param, &asset.Performance{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterPerformance(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryPerformances(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.QueryPerformancesParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", utils.AnyContext, "orchestrator.performance:QueryPerformances", param, &asset.QueryPerformancesResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryPerformances(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestHandlePerfConflictAfterTimeout(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := common.WithLastError(context.Background(), fabricTimeout)
	invocator := &mockedInvocator{}

	newPerf := &asset.NewPerformance{
		ComputeTaskKey: "taskUuid",
		MetricKey:      "metricUuid",
	}

	// register fail
	invocator.On(
		"Call",
		utils.AnyContext,
		"orchestrator.performance:RegisterPerformance",
		newPerf,
		&asset.Performance{},
	).Return(errors.NewError(errors.ErrConflict, "test"))

	// perf already registered
	param := &asset.QueryPerformancesParam{
		PageToken: "",
		PageSize:  1,
		Filter: &asset.PerformanceQueryFilter{
			ComputeTaskKey: newPerf.ComputeTaskKey,
			MetricKey:      newPerf.MetricKey,
		},
	}
	invocator.On(
		"Call",
		utils.AnyContext,
		"orchestrator.performance:QueryPerformances",
		param,
		&asset.QueryPerformancesResponse{},
	).Run(func(args mock.Arguments) {
		response := args.Get(3).(*asset.QueryPerformancesResponse)
		response.Performances = []*asset.Performance{
			{
				ComputeTaskKey: newPerf.ComputeTaskKey,
				MetricKey:      newPerf.MetricKey,
			},
		}
	}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterPerformance(ctx, newPerf)

	assert.NoError(t, err, "Query should pass")
}
