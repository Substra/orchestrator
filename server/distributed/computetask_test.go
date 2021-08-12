package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/common"
	"github.com/stretchr/testify/assert"
)

func TestComputeTaskAdapterImplementServer(t *testing.T) {
	adapter := NewComputeTaskAdapter()
	assert.Implements(t, (*asset.ComputeTaskServiceServer)(nil), adapter)
}

func TestRegisterTasks(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.RegisterTasksParam{}

	invocator.On("Call", AnyContext, "orchestrator.computetask:RegisterTasks", param, nil).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterTasks(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryTasks(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.QueryTasksParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", AnyContext, "orchestrator.computetask:QueryTasks", param, &asset.QueryTasksResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryTasks(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestHandleTasksConflictAfterTimeout(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := common.WithLastError(context.Background(), fabricTimeout)
	invocator := &mockedInvocator{}

	param := &asset.RegisterTasksParam{
		Tasks: []*asset.NewComputeTask{{}},
	}

	invocator.On("Call", AnyContext, "orchestrator.computetask:RegisterTasks", param, nil).Return(errors.ErrConflict)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterTasks(ctx, param)

	assert.NoError(t, err, "Registration should pass")
}

func TestHandleTasksBatchConflictAfterTimeout(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := common.WithLastError(context.Background(), fabricTimeout)
	invocator := &mockedInvocator{}

	param := &asset.RegisterTasksParam{
		Tasks: []*asset.NewComputeTask{{}, {}, {}},
	}

	invocator.On("Call", AnyContext, "orchestrator.computetask:RegisterTasks", param, nil).Return(errors.ErrConflict)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterTasks(ctx, param)

	// We cannot assume that ALL tasks have been registered.
	assert.Error(t, err, "Registration should fail because batch contains more than one task")
}
