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

package distributed

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/lib/asset"
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

// mockContract in a mock of contracts.ContractCollection
type mockContractCollection struct {
	mock.Mock
}

func (m *mockContractCollection) GetAllContracts() []contractapi.ContractInterface {
	args := m.Called()
	return args.Get(0).([]contractapi.ContractInterface)
}

func (m *mockContractCollection) IsEvaluateMethod(method string) bool {
	args := m.Called(method)
	return args.Get(0).(bool)
}

func TestContractInvocator(t *testing.T) {
	contract := &gateway.Contract{}
	checker := &mockContractCollection{}

	invocator := NewContractInvocator(contract, checker, []string{})

	assert.Implementsf(t, (*Invocator)(nil), invocator, "ContractInvocator should implements Invocator")
}

func TestParamWrapping(t *testing.T) {
	contract := &mockContract{}
	checker := &mockContractCollection{}

	invocator := NewContractInvocator(contract, checker, []string{})

	// Invocation param is a protoreflect.ProtoMessage
	param := &asset.ObjectivesQueryParam{PageToken: "uuid", PageSize: 20}
	wrapper, err := communication.Wrap(param)
	assert.NoError(t, err)

	// Which is serialized
	serializedInput, err := json.Marshal(wrapper)
	assert.NoError(t, err)
	// And converted to strings to match gateway contract
	expectedInput := []string{string(serializedInput)}

	// Response is also a wrapper
	response := &asset.ObjectivesQueryResponse{Objectives: []*asset.Objective{}, NextPageToken: "test"}
	wrappedResponse, err := communication.Wrap(response)
	assert.NoError(t, err)
	// Then serialized to match contractapi
	serializedResponse, err := json.Marshal(wrappedResponse)
	assert.NoError(t, err)

	// Here we use a submit where it should be an evaluate because evaluation is impossible to test due to limitations
	// from fabric.
	checker.On("IsEvaluateMethod", "org.substra.objective:QueryObjectives").Return(false)
	contract.On("SubmitTransaction", "org.substra.objective:QueryObjectives", expectedInput).Return(serializedResponse, nil)

	output := &asset.ObjectivesQueryResponse{}
	err = invocator.Call("org.substra.objective:QueryObjectives", param, output)
	assert.NoError(t, err)

	assert.Equal(t, "test", output.NextPageToken, "response should be properly unwrapped")
}

func TestNoOutput(t *testing.T) {
	contract := &mockContract{}
	checker := &mockContractCollection{}

	invocator := NewContractInvocator(contract, checker, []string{})

	expectedInput := getEmptyExpectedInput(t)

	checker.On("IsEvaluateMethod", "org.test:NoOutput").Return(false)
	contract.On("SubmitTransaction", "org.test:NoOutput", expectedInput).Return([]byte{}, nil)

	err := invocator.Call("org.test:NoOutput", nil, nil)
	assert.NoError(t, err)
}

func TestInvoke(t *testing.T) {
	contract := &mockContract{}
	checker := &mockContractCollection{}

	invocator := NewContractInvocator(contract, checker, []string{})
	expectedInput := getEmptyExpectedInput(t)

	checker.On("IsEvaluateMethod", "org.substra.some_contract:SomeMethod").Return(false)
	contract.On("SubmitTransaction", "org.substra.some_contract:SomeMethod", expectedInput).Return([]byte{}, nil)

	err := invocator.Call("org.substra.some_contract:SomeMethod", nil, nil)
	assert.NoError(t, err)
}

func getEmptyExpectedInput(t *testing.T) []string {

	// Invocation param is a protoreflect.ProtoMessage
	wrapper, err := communication.Wrap(nil)
	assert.NoError(t, err)

	// Which is serialized
	serializedInput, err := json.Marshal(wrapper)
	assert.NoError(t, err)

	// And converted to strings to match gateway contract
	return []string{string(serializedInput)}
}
