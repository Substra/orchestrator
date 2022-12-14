package adapters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/distributed/chaincode"
	"github.com/substra/orchestrator/server/distributed/interceptors"
	"github.com/substra/orchestrator/utils"
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
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.algo:RegisterAlgo", newObj, &asset.Algo{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterAlgo(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestGetAlgo(t *testing.T) {
	adapter := NewAlgoAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.GetAlgoParam{Key: "uuid"}

	invocator.On("Call", utils.AnyContext, "orchestrator.algo:GetAlgo", param, &asset.Algo{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.GetAlgo(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryAlgos(t *testing.T) {
	adapter := NewAlgoAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.QueryAlgosParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", utils.AnyContext, "orchestrator.algo:QueryAlgos", param, &asset.QueryAlgosResponse{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.QueryAlgos(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestHandleAlgoConflictAfterTimeout(t *testing.T) {
	adapter := NewAlgoAdapter()

	newObj := &asset.NewAlgo{
		Key: "uuid",
	}

	newCtx := commonInterceptors.WithLastError(context.Background(), FabricTimeout)
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.algo:RegisterAlgo", newObj, &asset.Algo{}).Return(errors.NewError(errors.ErrConflict, "test"))

	invocator.On("Call", utils.AnyContext, "orchestrator.algo:GetAlgo", &asset.GetAlgoParam{Key: newObj.Key}, &asset.Algo{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterAlgo(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestUpdateAlgo(t *testing.T) {
	adapter := NewAlgoAdapter()

	updatedA := &asset.UpdateAlgoParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated algo name",
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.algo:UpdateAlgo", updatedA, nil).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.UpdateAlgo(ctx, updatedA)
	assert.NoError(t, err, "Update should pass")
}
