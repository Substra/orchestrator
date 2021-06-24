package testing

import (
	"context"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/mock"
)

// MockedContext is a convenience mock of the OrchestrationTransactionContext
type MockedContext struct {
	mock.Mock
}

// SetContext is a mock
func (m *MockedContext) SetContext(ctx context.Context) {
}

// GetContext is a mock
func (m *MockedContext) GetContext() context.Context {
	return context.Background()
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

// GetProvider is a mock
func (m *MockedContext) GetProvider() service.DependenciesProvider {
	args := m.Called()
	return args.Get(0).(service.DependenciesProvider)
}

// GetDispatcher is a mock
func (m *MockedContext) GetDispatcher() event.Dispatcher {
	args := m.Called()
	return args.Get(0).(event.Dispatcher)
}
