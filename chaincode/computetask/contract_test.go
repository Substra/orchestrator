package computetask

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/ledger"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
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
	input := &asset.RegisterTasksParam{Tasks: tasks}
	wrapper, err := communication.Wrap(context.Background(), input)
	assert.NoError(t, err)
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(ledger.MockTransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterTasks", tasks, org).Return(nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	_, err = contract.RegisterTasks(ctx, wrapper)
	assert.NoError(t, err, "task registration should not fail")
}
