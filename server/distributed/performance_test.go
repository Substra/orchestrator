package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/common"
	"github.com/stretchr/testify/assert"
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

	invocator.On("Call", AnyContext, "orchestrator.performance:RegisterPerformance", param, &asset.Performance{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterPerformance(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestGetPerformance(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.GetComputeTaskPerformanceParam{}

	invocator.On("Call", AnyContext, "orchestrator.performance:GetComputeTaskPerformance", param, &asset.Performance{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.GetComputeTaskPerformance(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestHandlePerfConflictAfterTimeout(t *testing.T) {
	adapter := NewPerformanceAdapter()

	newCtx := common.WithLastError(context.Background(), fabricTimeout)
	invocator := &mockedInvocator{}

	param := &asset.NewPerformance{
		ComputeTaskKey: "uuid",
	}

	invocator.On("Call", AnyContext, "orchestrator.performance:RegisterPerformance", param, &asset.Performance{}).
		Return(errors.NewError(errors.ErrConflict, "test"))
	invocator.On(
		"Call",
		AnyContext,
		"orchestrator.performance:GetComputeTaskPerformance",
		&asset.GetComputeTaskPerformanceParam{ComputeTaskKey: param.ComputeTaskKey},
		&asset.Performance{},
	).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterPerformance(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
