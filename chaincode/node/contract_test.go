package node

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	testHelper "github.com/substrafoundation/substra-orchestrator/chaincode/testing"
	"github.com/substrafoundation/substra-orchestrator/lib/assets/node"
)

type MockedService struct {
	mock.Mock
}

func (m *MockedService) RegisterNode(n *node.Node) error {
	args := m.Called(n)
	return args.Error(0)
}

func (m *MockedService) GetNodes() ([]*node.Node, error) {
	args := m.Called()
	return args.Get(0).([]*node.Node), args.Error(1)
}

func mockFactory(mock node.API) func(c contractapi.TransactionContextInterface) (node.API, error) {
	return func(_ contractapi.TransactionContextInterface) (node.API, error) {
		return mock, nil
	}
}

func TestRegistration(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	org := "TestOrg"

	o := &node.Node{Id: org}
	mockService.On("RegisterNode", o).Return(nil).Once()

	sID := msp.SerializedIdentity{
		Mspid: org,
	}
	b, err := proto.Marshal(&sID)
	require.Nil(t, err, "SID marshal should not fail")

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(testHelper.MockedContext)
	ctx.On("GetStub").Return(stub).Once()

	err = contract.RegisterNode(ctx)
	assert.Nil(t, err, "node registration should not fail")
}

func TestQueryNodes(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	nodes := []*node.Node{
		{Id: "org1"},
		{Id: "org2"},
	}

	mockService.On("GetNodes").Return(nodes, nil).Once()

	ctx := new(testHelper.MockedContext)
	resp, err := contract.QueryNodes(ctx)
	assert.Nil(t, err, "querying nodes should not fail")
	assert.Len(t, resp, len(nodes), "query should return all nodes")
}
