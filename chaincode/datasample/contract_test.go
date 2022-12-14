package datasample

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/service"
)

func getMockedService(ctx *ledger.MockTransactionContext) *service.MockDataSampleAPI {
	mockService := new(service.MockDataSampleAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetDataSampleService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("GetContext").Return(context.Background())

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
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	datasamples := []*asset.DataSample{
		{Key: "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83"},
	}

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterDataSamples", newSamples, mspid).Return(datasamples, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err = contract.RegisterDataSamples(ctx, wrapper)
	assert.NoError(t, err, "Smart contract execution should not fail")
}

func TestUpdate(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	updateDataSample := &asset.UpdateDataSamplesParam{
		Keys:            []string{"4c67ad88-309a-48b4-8bc4-c2e2c1a87a83", "9eef1e88-951a-44fb-944a-c3dbd1d72d85"},
		DataManagerKeys: []string{"0b4b4466-9a81-4084-9bab-80939b78addd", "5067eb48-b29e-4a2d-81a0-82033a7d2ef8"},
	}
	wrapper, err := communication.Wrap(context.Background(), updateDataSample)
	assert.NoError(t, err)

	ctx := new(ledger.MockTransactionContext)

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
	filter := (*asset.DataSampleQueryFilter)(nil)

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("QueryDataSamples", &common.Pagination{Token: "", Size: 10}, filter).Return(datasamples, "", nil).Once()

	param := &asset.QueryDataSamplesParam{PageToken: "", PageSize: 10}
	wrapper, err := communication.Wrap(context.Background(), param)
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

	query := []string{
		"GetDataSample",
		"QueryDataSamples",
	}

	assert.Equal(t, query, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
