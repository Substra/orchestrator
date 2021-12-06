package metric

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *ledger.MockTransactionContext) *service.MockMetricAPI {
	mockService := new(service.MockMetricAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetMetricService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	addressable := &asset.Addressable{}
	newPerms := &asset.NewPermissions{}
	metadata := map[string]string{"test": "true"}

	mspid := "org"

	newObj := &asset.NewMetric{
		Key:            "uuid1",
		Name:           "Metric name",
		Description:    addressable,
		Address:        addressable,
		Metadata:       metadata,
		NewPermissions: newPerms,
	}
	wrapper, err := communication.Wrap(context.Background(), newObj)
	assert.NoError(t, err)

	o := &asset.Metric{}

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterMetric", newObj, mspid).Return(o, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err = contract.RegisterMetric(ctx, wrapper)
	assert.NoError(t, err)
}

func TestQueryMetrics(t *testing.T) {
	contract := &SmartContract{}

	metrics := []*asset.Metric{
		{Name: "test"},
		{Name: "test2"},
	}

	ctx := new(ledger.MockTransactionContext)
	service := getMockedService(ctx)
	service.On("QueryMetrics", &common.Pagination{Token: "", Size: 20}).Return(metrics, "", nil).Once()

	param := &asset.QueryMetricsParam{PageToken: "", PageSize: 20}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	wrapped, err := contract.QueryMetrics(ctx, wrapper)
	assert.NoError(t, err, "query should not fail")

	resp := new(asset.QueryMetricsResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Metrics, len(metrics), "query should return all metrics")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	queries := []string{
		"GetMetric",
		"QueryMetrics",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
