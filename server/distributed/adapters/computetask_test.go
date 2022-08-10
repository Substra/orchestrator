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

func TestComputeTaskAdapterImplementServer(t *testing.T) {
	adapter := NewComputeTaskAdapter()
	assert.Implements(t, (*asset.ComputeTaskServiceServer)(nil), adapter)
}

func TestRegisterTasks(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.RegisterTasksParam{}

	invocator.On("Call", utils.AnyContext, "orchestrator.computetask:RegisterTasks", param, &asset.RegisterTasksResponse{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterTasks(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryTasks(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.QueryTasksParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", utils.AnyContext, "orchestrator.computetask:QueryTasks", param, &asset.QueryTasksResponse{}).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.QueryTasks(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestHandleTasksConflictAfterTimeout(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := commonInterceptors.WithLastError(context.Background(), FabricTimeout)
	invocator := &chaincode.MockInvocator{}

	param := &asset.RegisterTasksParam{
		Tasks: []*asset.NewComputeTask{
			{
				Key: "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			},
		},
	}

	invocator.On("Call", utils.AnyContext, "orchestrator.computetask:RegisterTasks", param, &asset.RegisterTasksResponse{}).Return(errors.NewError(errors.ErrConflict, "test"))
	invocator.On("Call", utils.AnyContext, "orchestrator.computetask:GetTask", &asset.GetTaskParam{Key: "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"}, &asset.ComputeTask{}).
		Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterTasks(ctx, param)

	assert.NoError(t, err, "Registration should pass")
}

func TestHandleTasksBatchConflictAfterTimeout(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := commonInterceptors.WithLastError(context.Background(), FabricTimeout)
	invocator := &chaincode.MockInvocator{}

	param := &asset.RegisterTasksParam{
		Tasks: []*asset.NewComputeTask{{}, {}, {}},
	}

	invocator.On("Call", utils.AnyContext, "orchestrator.computetask:RegisterTasks", param, &asset.RegisterTasksResponse{}).Return(errors.NewError(errors.ErrConflict, "test"))

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterTasks(ctx, param)

	// We cannot assume that ALL tasks have been registered.
	assert.Error(t, err, "Registration should fail because batch contains more than one task")
}

func TestGetTaskInputAssets(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	param := &asset.GetTaskInputAssetsParam{ComputeTaskKey: "uuid"}

	inputs := []*asset.ComputeTaskInputAsset{
		{Identifier: "test"},
	}

	invocator.
		On("Call", utils.AnyContext, "orchestrator.computetask:GetTaskInputAssets", param, &asset.GetTaskInputAssetsResponse{}).
		Run(func(args mock.Arguments) {
			resp := args.Get(3).(*asset.GetTaskInputAssetsResponse)
			resp.Assets = inputs
		}).
		Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	response, err := adapter.GetTaskInputAssets(ctx, param)

	assert.NoError(t, err)
	assert.Equal(t, inputs, response.Assets)
}
