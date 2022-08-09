package computeplan

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
func getMockedService(ctx *ledger.MockTransactionContext) *service.MockComputePlanAPI {
	mockService := new(service.MockComputePlanAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetComputePlanService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	input := &asset.NewComputePlan{}
	wrapper, err := communication.Wrap(context.Background(), input)
	assert.NoError(t, err)
	output := &asset.ComputePlan{Key: "test"}
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterPlan", input, org).Return(output, nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	resp, err := contract.RegisterPlan(ctx, wrapper)
	assert.NoError(t, err, "plan registration should not fail")
	task := new(asset.ComputePlan)
	err = resp.Unwrap(task)
	assert.NoError(t, err)
	assert.Equal(t, task, output)
}

func TestApplyAction(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	input := &asset.ApplyPlanActionParam{Key: "test", Action: asset.ComputePlanAction_PLAN_ACTION_CANCELED}
	wrapper, err := communication.Wrap(context.Background(), input)
	assert.NoError(t, err)
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("ApplyPlanAction", input.Key, input.Action, org).Return(nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	err = contract.ApplyPlanAction(ctx, wrapper)
	assert.NoError(t, err, "plan action application should not fail")
}

func TestQueryComputePlans(t *testing.T) {
	contract := &SmartContract{}

	computePlans := []*asset.ComputePlan{
		{Tag: "test"},
		{Tag: "test2"},
	}

	filter := &asset.PlanQueryFilter{
		Owner: "owner",
	}

	ctx := new(ledger.MockTransactionContext)
	service := getMockedService(ctx)
	service.On("QueryPlans", &common.Pagination{Token: "", Size: 20}, filter).Return(computePlans, "", nil).Once()

	param := &asset.QueryPlansParam{Filter: filter, PageToken: "", PageSize: 20}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	wrapped, err := contract.QueryPlans(ctx, wrapper)
	assert.NoError(t, err, "query should not fail")
	resp := new(asset.QueryPlansResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Plans, len(computePlans), "query should return all compute plans")
}

func TestEvaluateTransactions(t *testing.T) {
	contract := &SmartContract{}

	queries := []string{
		"GetPlan",
		"QueryPlans",
	}

	assert.Equal(t, queries, contract.GetEvaluateTransactions(), "All non-commit transactions should be flagged")
}

func TestUpdate(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"
	updateComputePlanParam := &asset.UpdateComputePlanParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated compute plan name",
	}
	wrapper, err := communication.Wrap(context.Background(), updateComputePlanParam)
	assert.NoError(t, err)

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("UpdatePlan", updateComputePlanParam, mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()
	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.UpdatePlan(ctx, wrapper)
	assert.NoError(t, err, "Smart contract execution should not fail")
}
