package algo

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/mocks"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *mocks.TransactionContext) *service.MockAlgoAPI {
	mockService := new(service.MockAlgoAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetAlgoService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()
	ctx.On("SetRequestID", "").Once()
	ctx.On("GetContext").Return(context.Background())

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
		Category:       asset.AlgoCategory_ALGO_COMPOSITE,
		Description:    addressable,
		Algorithm:      addressable,
		Metadata:       metadata,
		NewPermissions: newPerms,
	}

	params, err := communication.Wrap(context.Background(), newObj)
	assert.NoError(t, err)

	a := &asset.Algo{}

	ctx := new(mocks.TransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterAlgo", newObj, mspid).Return(a, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err = contract.RegisterAlgo(ctx, params)
	assert.NoError(t, err)
}

func TestQueryAlgos(t *testing.T) {
	contract := &SmartContract{}

	algos := []*asset.Algo{
		{Name: "test", Category: asset.AlgoCategory_ALGO_SIMPLE},
		{Name: "test2", Category: asset.AlgoCategory_ALGO_SIMPLE},
	}

	filter := &asset.AlgoQueryFilter{
		Category: asset.AlgoCategory_ALGO_SIMPLE,
	}

	ctx := new(mocks.TransactionContext)
	service := getMockedService(ctx)
	service.On("QueryAlgos", &common.Pagination{Token: "", Size: 20}, filter).Return(algos, "", nil).Once()

	param := &asset.QueryAlgosParam{Filter: filter, PageToken: "", PageSize: 20}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	wrapped, err := contract.QueryAlgos(ctx, wrapper)
	assert.NoError(t, err, "query should not fail")
	resp := new(asset.QueryAlgosResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Algos, len(algos), "query should return all algos")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	queries := []string{
		"GetAlgo",
		"QueryAlgos",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
