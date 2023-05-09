package adapters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/distributed/chaincode"
	"github.com/substra/orchestrator/server/distributed/interceptors"
	"github.com/substra/orchestrator/utils"
)

func TestPerformanceAdapterImplementServer(t *testing.T) {
	adapter := NewPerformanceAdapter()
	assert.Implements(t, (*asset.PerformanceServiceServer)(nil), adapter)
}

func TestRegisterPerformance(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.NewPerformance{}

	invocator.On("Call", utils.AnyContext, "orchestrator.performance:RegisterPerformance", param, &asset.Performance{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterPerformance(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryPerformances(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.QueryPerformancesParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", utils.AnyContext, "orchestrator.performance:QueryPerformances", param, &asset.QueryPerformancesResponse{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.QueryPerformances(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestHandlePerfConflictAfterTimeout(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := commonInterceptors.WithLastError(context.Background(), FabricTimeout)
	invocator := &chaincode.MockInvocator{}

	newPerf := &asset.NewPerformance{
		ComputeTaskKey:              "taskUuid",
		ComputeTaskOutputIdentifier: "my_perf",
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
			},
		}
	}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterPerformance(ctx, newPerf)

	assert.NoError(t, err, "Query should pass")
}
