package algo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/service"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *ledger.MockTransactionContext) *service.MockAlgoAPI {
	mockService := new(service.MockAlgoAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetAlgoService").Return(mockService).Once()

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

	newObj := &asset.NewAlgo{
		Key:            "uuid1",
		Name:           "Algo name",
		Description:    addressable,
		Algorithm:      addressable,
		Metadata:       metadata,
		NewPermissions: newPerms,
	}

	params, err := communication.Wrap(context.Background(), newObj)
	assert.NoError(t, err)

	a := &asset.Algo{}

	ctx := new(ledger.MockTransactionContext)

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

	computePlanKey := uuid.NewString()

	algos := []*asset.Algo{
		{Name: "test"},
		{Name: "test2"},
	}

	filter := &asset.AlgoQueryFilter{
		ComputePlanKey: computePlanKey,
	}

	ctx := new(ledger.MockTransactionContext)
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

func TestUpdate(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	updateAlgoParam := &asset.UpdateAlgoParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated algo name",
	}
	wrapper, err := communication.Wrap(context.Background(), updateAlgoParam)
	assert.NoError(t, err)

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("UpdateAlgo", updateAlgoParam, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.UpdateAlgo(ctx, wrapper)
	assert.NoError(t, err, "Smart contract execution should not fail")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	queries := []string{
		"GetAlgo",
		"QueryAlgos",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
