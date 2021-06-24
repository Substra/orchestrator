package computeplan

import (
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *testHelper.MockedContext) *service.MockComputePlanService {
	mockService := new(service.MockComputePlanService)

	provider := new(service.MockServiceProvider)
	provider.On("GetComputePlanService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	input := &asset.NewComputePlan{}
	wrapper, err := communication.Wrap(input)
	assert.NoError(t, err)
	output := &asset.ComputePlan{Key: "test"}
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(testHelper.MockedContext)

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
	wrapper, err := communication.Wrap(input)
	assert.NoError(t, err)
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("ApplyPlanAction", input.Key, input.Action, org).Return(nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	err = contract.ApplyPlanAction(ctx, wrapper)
	assert.NoError(t, err, "plan action application should not fail")
}
