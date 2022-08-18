package ledger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetProvider(t *testing.T) {
	stub := new(testHelper.MockedStub)

	ctx := NewContext()
	ctx.SetContext(context.Background())
	ctx.SetStub(stub)

	stub.On("GetTxTimestamp").Once().Return(timestamppb.Now(), nil)
	stub.On("GetChannelID").Once().Return("testChannel", nil)

	provider, err := ctx.GetProvider()
	assert.NoError(t, err)
	assert.Implements(t, (*service.DependenciesProvider)(nil), provider, "GetProvider should return a service provider")
}

func TestAfterTransactionHook(t *testing.T) {
	ctx := NewContext()
	ctx.SetContext(context.Background())

	dispatcher := new(MockEventDispatcher)
	ctx.dispatcher = dispatcher

	dispatcher.On("Dispatch").Once().Return(nil)

	err := AfterTransactionHook(ctx, "whatever")
	assert.NoError(t, err)
}

func TestIsEvaluateTransaction(t *testing.T) {
	evalFuncs := []string{"GetAllOrganizations"}

	assert.True(t, IsEvaluateTransaction("orchestrator.organization:GetAllOrganizations", evalFuncs))
	assert.True(t, IsEvaluateTransaction("GetAllOrganizations", evalFuncs))

	assert.False(t, IsEvaluateTransaction("orchestrator.organization:RegisterOrganization", evalFuncs))
	assert.False(t, IsEvaluateTransaction("RegisterOrganization", evalFuncs))
}
