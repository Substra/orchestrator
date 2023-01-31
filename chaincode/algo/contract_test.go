package function

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
func getMockedService(ctx *ledger.MockTransactionContext) *service.MockFunctionAPI {
	mockService := new(service.MockFunctionAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetFunctionService").Return(mockService).Once()

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

	newObj := &asset.NewFunction{
		Key:            "uuid1",
		Name:           "Function name",
		Description:    addressable,
		Function:      addressable,
		Metadata:       metadata,
		NewPermissions: newPerms,
	}

	params, err := communication.Wrap(context.Background(), newObj)
	assert.NoError(t, err)

	a := &asset.Function{}

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterFunction", newObj, mspid).Return(a, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err = contract.RegisterFunction(ctx, params)
	assert.NoError(t, err)
}

func TestQueryFunctions(t *testing.T) {
	contract := &SmartContract{}

	computePlanKey := uuid.NewString()

	functions := []*asset.Function{
		{Name: "test"},
		{Name: "test2"},
	}

	filter := &asset.FunctionQueryFilter{
		ComputePlanKey: computePlanKey,
	}

	ctx := new(ledger.MockTransactionContext)
	service := getMockedService(ctx)
	service.On("QueryFunctions", &common.Pagination{Token: "", Size: 20}, filter).Return(functions, "", nil).Once()

	param := &asset.QueryFunctionsParam{Filter: filter, PageToken: "", PageSize: 20}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	wrapped, err := contract.QueryFunctions(ctx, wrapper)
	assert.NoError(t, err, "query should not fail")
	resp := new(asset.QueryFunctionsResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Functions, len(functions), "query should return all functions")
}

func TestUpdate(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	updateFunctionParam := &asset.UpdateFunctionParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated function name",
	}
	wrapper, err := communication.Wrap(context.Background(), updateFunctionParam)
	assert.NoError(t, err)

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("UpdateFunction", updateFunctionParam, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.UpdateFunction(ctx, wrapper)
	assert.NoError(t, err, "Smart contract execution should not fail")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	queries := []string{
		"GetFunction",
		"QueryFunctions",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}
