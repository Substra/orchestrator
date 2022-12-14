package computetask

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/chaincode/communication"
	"github.com/substra/orchestrator/chaincode/ledger"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/service"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *ledger.MockTransactionContext) *service.MockComputeTaskAPI {
	mockService := new(service.MockComputeTaskAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetComputeTaskService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider, nil).Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	tasks := []*asset.NewComputeTask{{}, {}}
	registeredTasks := []*asset.ComputeTask{{}, {}}
	input := &asset.RegisterTasksParam{Tasks: tasks}
	wrapper, err := communication.Wrap(context.Background(), input)
	assert.NoError(t, err)
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterTasks", tasks, org).Return(registeredTasks, nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	_, err = contract.RegisterTasks(ctx, wrapper)
	assert.NoError(t, err, "task registration should not fail")
}

func TestGetTaskInputAssets(t *testing.T) {
	contract := &SmartContract{}

	param := &asset.GetTaskInputAssetsParam{ComputeTaskKey: "uuid"}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	stub := new(testHelper.MockedStub)

	ctx := new(ledger.MockTransactionContext)

	inputs := []*asset.ComputeTaskInputAsset{
		{Identifier: "test"},
	}

	service := getMockedService(ctx)
	service.On("GetInputAssets", "uuid").Return(inputs, nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	w, err := contract.GetTaskInputAssets(ctx, wrapper)
	assert.NoError(t, err)

	resp := new(asset.GetTaskInputAssetsResponse)
	err = w.Unwrap(resp)
	assert.NoError(t, err)
	assert.Equal(t, inputs, resp.Assets)

	service.AssertExpectations(t)
}
