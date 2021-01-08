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
	"github.com/owkin/orchestrator/lib/orchestration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedService struct {
	mock.Mock
}

func (m *MockedService) RegisterObjective(o *assets.NewObjective, owner string) (*assets.Objective, error) {
	args := m.Called(o, owner)
	return args.Get(0).(*assets.Objective), args.Error(1)
}

func (m *MockedService) GetObjective(key string) (*assets.Objective, error) {
	args := m.Called(key)
	return args.Get(0).(*assets.Objective), args.Error(1)
}

func (m *MockedService) GetObjectives() ([]*assets.Objective, error) {
	args := m.Called()
	return args.Get(0).([]*assets.Objective), args.Error(1)
}

func mockFactory(mock orchestration.ObjectiveAPI) func(c contractapi.TransactionContextInterface) (orchestration.ObjectiveAPI, error) {
	return func(_ contractapi.TransactionContextInterface) (orchestration.ObjectiveAPI, error) {
		return mock, nil
	}
}

func TestRegistration(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	addressable := &assets.Addressable{}
	testDataset := &assets.Dataset{}
	newPerms := &assets.NewPermissions{}
	metadata := map[string]string{"test": "true"}

	mspid := "org"

	newObj := &assets.NewObjective{
		Key:            "uuid1",
		Name:           "Objective name",
		Description:    addressable,
		MetricsName:    "metrics name",
		Metrics:        addressable,
		TestDataset:    testDataset,
		Metadata:       metadata,
		NewPermissions: &assets.NewPermissions{},
	}

	o := &assets.Objective{}
	mockService.On("RegisterObjective", newObj, mspid).Return(o, nil).Once()

	ctx := new(testHelper.MockedContext)
	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	contract.RegisterObjective(
		ctx,
		"uuid1",
		"Objective name",
		addressable,
		"metrics name",
		addressable,
		testDataset,
		metadata,
		newPerms,
	)
}

func TestQueryObjectives(t *testing.T) {
	mockService := new(MockedService)
	contract := &SmartContract{
		serviceFactory: mockFactory(mockService),
	}

	objectives := []*assets.Objective{
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
	assert.Implements(t, (*orchestration.ObjectiveAPI)(nil), service, "service should implements objective API")
}
