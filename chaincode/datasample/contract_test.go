package datasample

import (
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
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
	newSamples := []*asset.NewDataSample{
		{
			Key:             "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
			DataManagerKeys: []string{"0b4b4466-9a81-4084-9bab-80939b78addd"},
			TestOnly:        false,
		},
	}
	param := &asset.RegisterDataSamplesParam{Samples: newSamples}
	wrapper, err := communication.Wrap(param)
	assert.NoError(t, err)

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("RegisterDataSamples", newSamples, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.RegisterDataSamples(ctx, wrapper)
	assert.NoError(t, err, "Smart contract execution should not fail")
}

func TestUpdate(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	updateDataSample := &asset.UpdateDataSamplesParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		DataManagerKeys: []string{"0b4b4466-9a81-4084-9bab-80939b78addd", "5067eb48-b29e-4a2d-81a0-82033a7d2ef8"},
	}
	wrapper, err := communication.Wrap(updateDataSample)
	assert.NoError(t, err)

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("UpdateDataSamples", updateDataSample, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.UpdateDataSamples(ctx, wrapper)
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
	service.On("QueryDataSamples", &common.Pagination{Token: "", Size: 10}).Return(datasamples, "", nil).Once()

	param := &asset.QueryDataSamplesParam{PageToken: "", PageSize: 10}
	wrapper, err := communication.Wrap(param)
	assert.NoError(t, err)

	wrapped, err := contract.QueryDataSamples(ctx, wrapper)
	assert.NoError(t, err, "Query should not fail")
	resp := new(asset.QueryDataSamplesResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.Len(t, resp.DataSamples, len(datasamples), "Query should return all datasamples")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	query := []string{"QueryDataSamples"}

	assert.Equal(t, query, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
