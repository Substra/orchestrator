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

package objective

import (
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/assets/dataset"
	"github.com/owkin/orchestrator/lib/assets/objective"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedService struct {
	mock.Mock
}

func (m *MockedService) RegisterObjective(o *objective.Objective) error {
	args := m.Called(o)
	return args.Error(0)
}

func (m *MockedService) GetObjective(key string) (*objective.Objective, error) {
	args := m.Called(key)
	return args.Get(0).(*objective.Objective), args.Error(1)
}

func (m *MockedService) GetObjectives() ([]*objective.Objective, error) {
	args := m.Called()
	return args.Get(0).([]*objective.Objective), args.Error(1)
}

func mockFactory(mock objective.API) func(c contractapi.TransactionContextInterface) (objective.API, error) {
	return func(_ contractapi.TransactionContextInterface) (objective.API, error) {
		return mock, nil
	}
}

func TestRegistration(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	description := &assets.Addressable{}
	metrics := &assets.Addressable{}
	testDataset := &dataset.Dataset{}
	permissions := &assets.Permissions{}
	metadata := map[string]string{"test": "true"}

	o := &objective.Objective{
		Key:         "uuid1",
		Name:        "Objective name",
		Description: description,
		MetricsName: "metrics name",
		Metrics:     metrics,
		TestDataset: testDataset,
		Metadata:    metadata,
		Permissions: permissions,
	}
	mockService.On("RegisterObjective", o).Return(nil).Once()

	ctx := new(testHelper.MockedContext)
	contract.RegisterObjective(
		ctx,
		"uuid1",
		"Objective name",
		description,
		"metrics name",
		metrics,
		testDataset,
		metadata,
		permissions,
	)
}

func TestQueryObjectives(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	objectives := []*objective.Objective{
		{Name: "test"},
		{Name: "test2"},
	}

	mockService.On("GetObjectives").Return(objectives, nil).Once()

	ctx := new(testHelper.MockedContext)
	r, err := contract.QueryObjectives(ctx)
	assert.Nil(t, err, "query should not fail")
	assert.Len(t, r, len(objectives), "query should return all objectives")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := NewSmartContract()

	queries := []string{
		"QueryObjectives",
		"QueryLeaderboard",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}

func TestGetServiceFromContext(t *testing.T) {
	context := testHelper.MockedContext{}
	context.On("GetStub").Return(&testHelper.MockedStub{}).Once()

	service, err := getServiceFromContext(&context)

	assert.Nil(t, err, "Creating service should not fail")
	assert.Implements(t, (*objective.API)(nil), service, "service should implements objective API")
}
