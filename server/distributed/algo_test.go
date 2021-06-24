package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestAlgoAdapterImplementServer(t *testing.T) {
	adapter := NewAlgoAdapter()
	assert.Implementsf(t, (*asset.AlgoServiceServer)(nil), adapter, "AlgoAdapter should implements AlgoServiceServer")
}

func TestRegisterAlgo(t *testing.T) {
	adapter := NewAlgoAdapter()

	newObj := &asset.NewAlgo{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", "orchestrator.algo:RegisterAlgo", newObj, &asset.Algo{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterAlgo(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestGetAlgo(t *testing.T) {
	adapter := NewAlgoAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.GetAlgoParam{Key: "uuid"}

	invocator.On("Call", "orchestrator.algo:GetAlgo", param, &asset.Algo{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.GetAlgo(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryAlgos(t *testing.T) {
	adapter := NewAlgoAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.QueryAlgosParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", "orchestrator.algo:QueryAlgos", param, &asset.QueryAlgosResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryAlgos(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
