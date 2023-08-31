package failurereport

import (
	"context"
	"github.com/substra/orchestrator/lib/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
)

func getMockedService(ctx *ledger.MockTransactionContext) *service.MockFailureReportAPI {
	mockService := new(service.MockFailureReportAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetFailureReportService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestRegisterFailureReport(t *testing.T) {
	contract := &SmartContract{}
	mspid := "org"

	newFailureReport := &asset.NewFailureReport{
		AssetKey:    "taskUUID",
		AssetType:   asset.FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
		LogsAddress: &asset.Addressable{},
	}
	wrapper, err := communication.Wrap(context.Background(), newFailureReport)
	assert.NoError(t, err)

	ctx := new(ledger.MockTransactionContext)

	mockService := getMockedService(ctx)
	mockService.On("RegisterFailureReport", newFailureReport, mspid).Return(&asset.FailureReport{}, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err = contract.RegisterFailureReport(ctx, wrapper)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
	stub.AssertExpectations(t)
}

func TestGetFailureReport(t *testing.T) {
	contract := &SmartContract{}

	param := &asset.GetFailureReportParam{
		AssetKey: "uuid",
	}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	ctx := new(ledger.MockTransactionContext)

	failureReport := &asset.FailureReport{
		AssetKey:  param.AssetKey,
		AssetType: asset.FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
	}
	mockService := getMockedService(ctx)
	mockService.On("GetFailureReport", "uuid").Return(failureReport, nil).Once()

	wrapped, err := contract.GetFailureReport(ctx, wrapper)
	assert.NoError(t, err)

	resp := new(asset.FailureReport)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.AssetKey, param.AssetKey)

	mockService.AssertExpectations(t)
}
