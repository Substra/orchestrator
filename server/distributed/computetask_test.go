package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
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

	invocator.On("Call", "orchestrator.computetask:RegisterTasks", param, nil).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterTasks(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryTasks(t *testing.T) {
	adapter := NewComputeTaskAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.QueryTasksParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", "orchestrator.computetask:QueryTasks", param, &asset.QueryTasksResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryTasks(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
