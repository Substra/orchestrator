package node

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *ledger.MockTransactionContext) *service.MockNodeAPI {
	mockService := new(service.MockNodeAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetNodeService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	o := &asset.Node{Id: org}
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterNode", org).Return(o, nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	wrapper, err := communication.Wrap(context.Background(), nil)
	require.NoError(t, err)

	wrapped, err := contract.RegisterNode(ctx, wrapper)
	assert.NoError(t, err, "node registration should not fail")
	node := new(asset.Node)
	err = wrapped.Unwrap(node)
	assert.NoError(t, err)
	assert.Equal(t, node, o)
}

func TestGetAllNodes(t *testing.T) {
	contract := &SmartContract{}

	nodes := []*asset.Node{
		{Id: "org1"},
		{Id: "org2"},
	}

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("GetAllNodes").Return(nodes, nil).Once()

	wrapper, err := communication.Wrap(context.Background(), nil)
	require.NoError(t, err)

	wrapped, err := contract.GetAllNodes(ctx, wrapper)
	assert.NoError(t, err, "querying nodes should not fail")
	queryResult := new(asset.GetAllNodesResponse)
	err = wrapped.Unwrap(queryResult)
	assert.NoError(t, err)
	assert.Len(t, queryResult.Nodes, len(nodes), "query should return all nodes")
}
