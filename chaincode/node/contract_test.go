package node

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/mock"
	"github.com/substrafoundation/substra-orchestrator/lib/assets/node"
)

type MockedContext struct {
	mock.Mock
}

func (m *MockedContext) GetStub() shim.ChaincodeStubInterface {
	args := m.Called()
	return args.Get(0).(shim.ChaincodeStubInterface)
}

func (m *MockedContext) GetClientIdentity() cid.ClientIdentity {
	args := m.Called()
	return args.Get(0).(cid.ClientIdentity)
}

type MockedService struct {
	mock.Mock
}

func (m *MockedService) RegisterNode(n *node.Node) error {
	args := m.Called(n)
	return args.Error(0)
}

func mockFactory(mock node.Manager) func(c contractapi.TransactionContextInterface) (node.Manager, error) {
	return func(_ contractapi.TransactionContextInterface) (node.Manager, error) {
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

	ctx := new(MockedContext)
	contract.RegisterNode(ctx, "uuid1")
}
