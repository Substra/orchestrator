package dataset

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/mocks"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *mocks.TransactionContext) *service.MockDatasetAPI {
	mockService := new(service.MockDatasetAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetDatasetService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("SetRequestID", "").Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestEvaluateTransactions(t *testing.T) {
	contract := NewSmartContract()

	queries := []string{
		"GetDataset",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}

func TestGetDataset(t *testing.T) {
	contract := &SmartContract{}
	var key = "test"

	dataset := &asset.Dataset{
		DataManager: &asset.DataManager{
			Key: key,
		},
	}

	ctx := new(mocks.TransactionContext)

	service := getMockedService(ctx)
	service.On("GetDataset", key).Return(dataset, nil).Once()

	param := &asset.GetDatasetParam{Key: key}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	wrapped, err := contract.GetDataset(ctx, wrapper)
	assert.NoError(t, err, "Query should not fail")
	resp := new(asset.Dataset)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.EqualValuesf(t, resp.DataManager.Key, key, "Query should return dataset key")
}
