// Copyright 2020 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
