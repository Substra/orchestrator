package node

import (
	"testing"

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *testHelper.MockedContext) *service.MockNodeService {
	mockService := new(service.MockNodeService)

	provider := new(service.MockServiceProvider)
	provider.On("GetNodeService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	o := &asset.Node{Id: org}
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("RegisterNode", org).Return(o, nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	wrapped, err := contract.RegisterNode(ctx)
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

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("GetAllNodes").Return(nodes, nil).Once()

	wrapped, err := contract.GetAllNodes(ctx)
	assert.NoError(t, err, "querying nodes should not fail")
	queryResult := new(asset.GetAllNodesResponse)
	err = wrapped.Unwrap(queryResult)
	assert.NoError(t, err)
	assert.Len(t, queryResult.Nodes, len(nodes), "query should return all nodes")
}
