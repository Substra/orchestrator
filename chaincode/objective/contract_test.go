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

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/orchestration"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *testHelper.MockedContext) *orchestration.MockObjectiveService {
	mockService := new(orchestration.MockObjectiveService)

	provider := new(orchestration.MockServiceProvider)
	provider.On("GetObjectiveService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

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
		NewPermissions: newPerms,
	}

	o := &assets.Objective{}

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("RegisterObjective", newObj, mspid).Return(o, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	contract.RegisterObjective(ctx, newObj)
}

func TestQueryObjectives(t *testing.T) {
	contract := &SmartContract{}

	objectives := []*assets.Objective{
		{Name: "test"},
		{Name: "test2"},
	}

	ctx := new(testHelper.MockedContext)
	service := getMockedService(ctx)
	service.On("GetObjectives", &common.Pagination{Token: "", Size: 20}).Return(objectives, "", nil).Once()

	param := &assets.ObjectivesQueryParam{PageToken: "", PageSize: 20}

	resp, err := contract.QueryObjectives(ctx, param)
	assert.NoError(t, err, "query should not fail")
	assert.Len(t, resp.Objectives, len(objectives), "query should return all objectives")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	queries := []string{
		"QueryObjective",
		"QueryObjectives",
		"QueryLeaderboard",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
