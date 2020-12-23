package node

import (
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/mock"
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

	o := &node.Node{Id: "uuid1"}
	mockService.On("RegisterNode", o).Return(nil).Once()

	ctx := new(testHelper.MockedContext)
	contract.RegisterNode(ctx, "uuid1")
}
