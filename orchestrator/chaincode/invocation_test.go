// Copyright 2021 Owkin Inc.
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

package chaincode

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockContract in a mock of gateway.Contract
type mockContract struct {
	mock.Mock
}

func (m *mockContract) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockContract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	mockArgs := m.Called(name, args)
	return mockArgs.Get(0).([]byte), mockArgs.Error(1)
}

func (m *mockContract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	mockArgs := m.Called(name, args)
	return mockArgs.Get(0).([]byte), mockArgs.Error(1)
}

func (m *mockContract) CreateTransaction(name string, opts ...gateway.TransactionOption) (*gateway.Transaction, error) {
	args := m.Called(name, opts)
	return args.Get(0).(*gateway.Transaction), args.Error(1)
}

func (m *mockContract) RegisterEvent(eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	args := m.Called(eventFilter)
	return args.Get(0).(fab.Registration), args.Get(1).(<-chan *fab.CCEvent), args.Error(2)
}

func (m *mockContract) Unregister(registration fab.Registration) {
	m.Called(registration)
}

func TestContractInvocator(t *testing.T) {
	contract := &gateway.Contract{}

	invocator := NewContractInvocator(contract)

	assert.Implementsf(t, (*Invocator)(nil), invocator, "ContractInvocator should implements Invocator")
}

func TestParamWrapping(t *testing.T) {
	contract := &mockContract{}

	invocator := &ContractInvocator{contract: contract}

	// Invocation param is a protoreflect.ProtoMessage
	param := &assets.ObjectivesQueryParam{PageToken: "uuid", PageSize: 20}

	// Which is serialized
	serializedInput, err := json.Marshal(param)
	assert.NoError(t, err)
	// And converted to strings to match gateway contract
	expectedInput := []string{string(serializedInput)}

	// Response is also a protoreflect.ProtoMessage
	response := &assets.ObjectivesQueryResponse{Objectives: []*assets.Objective{}, NextPageToken: "test"}
	// Then serialized to match contractapi
	serializedResponse, err := json.Marshal(response)
	assert.NoError(t, err)

	contract.On("SubmitTransaction", "org.substra.objective:QueryObjectives", expectedInput).Return(serializedResponse, nil)

	output := &assets.ObjectivesQueryResponse{}
	invocator.Invoke("org.substra.objective:QueryObjectives", param, output)

	assert.Equal(t, "test", output.NextPageToken, "response should be properly unwrapped")
}
