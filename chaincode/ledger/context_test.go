package ledger

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

func TestGetProvider(t *testing.T) {
	ctx := NewContext()
	ctx.SetContext(context.Background())
	assert.Implements(t, (*service.DependenciesProvider)(nil), ctx.GetProvider(), "GetProvider should return a service provider")
}

func TestAfterTransactionHook(t *testing.T) {
	ctx := NewContext()
	ctx.SetContext(context.Background())

	dispatcher := new(event.MockDispatcher)
	ctx.dispatcher = dispatcher

	dispatcher.On("Dispatch").Once().Return(nil)

	err := AfterTransactionHook(ctx, "whatever")
	assert.NoError(t, err)
}

func TestIsEvaluateTransaction(t *testing.T) {
	evalFuncs := []string{"GetAllNodes"}

	assert.True(t, IsEvaluateTransaction("orchestrator.node:GetAllNodes", evalFuncs))
	assert.True(t, IsEvaluateTransaction("GetAllNodes", evalFuncs))

	assert.False(t, IsEvaluateTransaction("orchestrator.node:RegisterNodes", evalFuncs))
	assert.False(t, IsEvaluateTransaction("RegisterNodes", evalFuncs))
}
