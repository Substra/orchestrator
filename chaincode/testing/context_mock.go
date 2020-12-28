package testing

import (
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/mock"
)

// MockedContext is a convenience mock of the Transactioncontextinterface from contractapi
type MockedContext struct {
	mock.Mock
}

// GetStub is a mock
func (m *MockedContext) GetStub() shim.ChaincodeStubInterface {
	args := m.Called()
	return args.Get(0).(shim.ChaincodeStubInterface)
}

// GetClientIdentity is a mock
func (m *MockedContext) GetClientIdentity() cid.ClientIdentity {
	args := m.Called()
	return args.Get(0).(cid.ClientIdentity)
}
