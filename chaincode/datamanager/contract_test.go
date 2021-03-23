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

package datamanager

import (
	"testing"

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	query := []string{"QueryDataManager", "QueryDataManagers"}

	assert.Equal(t, query, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}

func getMockedService(ctx *testHelper.MockedContext) *service.MockDataManagerService {
	mockService := new(service.MockDataManagerService)

	provider := new(service.MockServiceProvider)
	provider.On("GetDataManagerService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestQueryDataManagers(t *testing.T) {
	contract := &SmartContract{}

	datamanagers := []*asset.DataManager{
		{Key: "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		{Key: "9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
	}

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("GetDataManagers", &common.Pagination{Token: "", Size: 10}).Return(datamanagers, "", nil).Once()

	param := &asset.DataManagersQueryParam{PageToken: "", PageSize: 10}

	resp, err := contract.QueryDataManagers(ctx, param)
	assert.NoError(t, err, "Query should not fail")
	assert.Len(t, resp.DataManagers, len(datamanagers), "Query should return all datasamples")
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	addressable := &asset.Addressable{}
	newPerms := &asset.NewPermissions{}
	metadata := map[string]string{"test": "true"}

	mspid := "org"

	newObj := &asset.NewDataManager{
		Key:            "uuid1",
		Name:           "Datamanager name",
		Description:    addressable,
		Metadata:       metadata,
		NewPermissions: newPerms,
		ObjectiveKey:   "uuid2",
		Opener:         addressable,
		Type:           "test",
	}

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("RegisterDataManager", newObj, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err := contract.RegisterDataManager(ctx, newObj)
	assert.NoError(t, err)
}
