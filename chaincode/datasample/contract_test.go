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

package datasample

import (
	"testing"

	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

func getMockedService(ctx *testHelper.MockedContext) *service.MockDataSampleService {
	mockService := new(service.MockDataSampleService)

	provider := new(service.MockServiceProvider)
	provider.On("GetDataSampleService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	newDataSample := &asset.NewDataSample{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		DataManagerKeys: []string{"0b4b4466-9a81-4084-9bab-80939b78addd"},
		TestOnly:        false,
	}

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("RegisterDataSample", newDataSample, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err := contract.RegisterDataSample(ctx, newDataSample)
	assert.NoError(t, err, "Smart contract execution should not fail")
}

func TestUpdate(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	updateDataSample := &asset.DataSampleUpdateParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		DataManagerKeys: []string{"0b4b4466-9a81-4084-9bab-80939b78addd", "5067eb48-b29e-4a2d-81a0-82033a7d2ef8"},
	}

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("UpdateDataSample", updateDataSample, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err := contract.UpdateDataSample(ctx, updateDataSample)
	assert.NoError(t, err, "Smart contract execution should not fail")
}

func TestQueryDataSamples(t *testing.T) {
	contract := &SmartContract{}

	datasamples := []*asset.DataSample{
		{Key: "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
		{Key: "9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
	}

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("GetDataSamples", &common.Pagination{Token: "", Size: 10}).Return(datasamples, "", nil).Once()

	param := &asset.DataSamplesQueryParam{PageToken: "", PageSize: 10}

	resp, err := contract.QueryDataSamples(ctx, param)
	assert.NoError(t, err, "Query should not fail")
	assert.Len(t, resp.DataSamples, len(datasamples), "Query should return all datasamples")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	query := []string{"QueryDataSamples"}

	assert.Equal(t, query, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
