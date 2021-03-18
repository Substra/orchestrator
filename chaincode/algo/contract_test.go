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

package algo

import (
	"testing"

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *testHelper.MockedContext) *service.MockAlgoService {
	mockService := new(service.MockAlgoService)

	provider := new(service.MockServiceProvider)
	provider.On("GetAlgoService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	addressable := &asset.Addressable{}
	newPerms := &asset.NewPermissions{}
	metadata := map[string]string{"test": "true"}

	mspid := "org"

	newObj := &asset.NewAlgo{
		Key:            "uuid1",
		Name:           "Algo name",
		Category:       asset.AlgoCategory_COMPOSITE,
		Description:    addressable,
		Algorithm:      addressable,
		Metadata:       metadata,
		NewPermissions: newPerms,
	}

	a := &asset.Algo{}

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("RegisterAlgo", newObj, mspid).Return(a, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err := contract.RegisterAlgo(ctx, newObj)
	assert.NoError(t, err)
}

func TestQueryAlgos(t *testing.T) {
	contract := &SmartContract{}

	algos := []*asset.Algo{
		{Name: "test"},
		{Name: "test2"},
	}

	ctx := new(testHelper.MockedContext)
	service := getMockedService(ctx)
	service.On("GetAlgos", &common.Pagination{Token: "", Size: 20}).Return(algos, "", nil).Once()

	param := &asset.AlgosQueryParam{PageToken: "", PageSize: 20}

	resp, err := contract.QueryAlgos(ctx, param)
	assert.NoError(t, err, "query should not fail")
	assert.Len(t, resp.Algos, len(algos), "query should return all algos")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	queries := []string{
		"QueryAlgo",
		"QueryAlgos",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
