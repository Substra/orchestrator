package info

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/mocks"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryVersion(t *testing.T) {
	contract := &SmartContract{}

	o := &asset.QueryVersionResponse{
		Chaincode:    "dev",
		Orchestrator: "",
	}
	ctx := new(mocks.TransactionContext)

	ctx.On("SetRequestID", "").Once()
	ctx.On("GetContext").Return(context.Background())

	wrapper, err := communication.Wrap(context.Background(), nil)
	require.NoError(t, err)

	wrapped, err := contract.QueryVersion(ctx, wrapper)
	assert.NoError(t, err, "QueryVersion should not fail")
	version := new(asset.QueryVersionResponse)
	err = wrapped.Unwrap(version)
	assert.NoError(t, err)
	assert.Equal(t, o.Chaincode, version.Chaincode)
	assert.Equal(t, o.Orchestrator, version.Orchestrator)
}
