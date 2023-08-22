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

func TestFunctionAdapterImplementServer(t *testing.T) {
	adapter := NewFunctionAdapter()
	assert.Implementsf(t, (*asset.FunctionServiceServer)(nil), adapter, "FunctionAdapter should implements FunctionServiceServer")
}

func TestRegisterFunction(t *testing.T) {
	adapter := NewFunctionAdapter()

	newObj := &asset.NewFunction{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.function:RegisterFunction", newObj, &asset.Function{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterFunction(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestGetFunction(t *testing.T) {
	adapter := NewFunctionAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.GetFunctionParam{Key: "uuid"}

	invocator.On("Call", utils.AnyContext, "orchestrator.function:GetFunction", param, &asset.Function{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.GetFunction(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryFunctions(t *testing.T) {
	adapter := NewFunctionAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.QueryFunctionsParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", utils.AnyContext, "orchestrator.function:QueryFunctions", param, &asset.QueryFunctionsResponse{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.QueryFunctions(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestHandleFunctionConflictAfterTimeout(t *testing.T) {
	adapter := NewFunctionAdapter()

	newObj := &asset.NewFunction{
		Key: "uuid",
	}

	newCtx := commonInterceptors.WithLastError(context.Background(), FabricTimeout)
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.function:RegisterFunction", newObj, &asset.Function{}).Return(errors.NewError(errors.ErrConflict, "test"))

	invocator.On("Call", utils.AnyContext, "orchestrator.function:GetFunction", &asset.GetFunctionParam{Key: newObj.Key}, &asset.Function{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterFunction(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestUpdateFunction(t *testing.T) {
	adapter := NewFunctionAdapter()

	updatedA := &asset.UpdateFunctionParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated function name",
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.function:UpdateFunction", updatedA, nil).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.UpdateFunction(ctx, updatedA)
	assert.NoError(t, err, "Update should pass")
}

func TestUpdateFunctionStatus(t *testing.T) {
	adapter := NewFunctionAdapter()

	updatedA := &asset.UpdateFunctionStatusParam{
		Key:    "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.function:UpdateFunctionStatus", updatedA, nil).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.UpdateFunctionStatus(ctx, updatedA)
	assert.NoError(t, err, "Update should pass")
}
