package computetask

import (
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *testHelper.MockedContext) *service.MockComputeTaskAPI {
	mockService := new(service.MockComputeTaskAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetComputeTaskService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestRegistration(t *testing.T) {
	contract := &SmartContract{}

	org := "TestOrg"
	tasks := []*asset.NewComputeTask{{}, {}}
	input := &asset.RegisterTasksParam{Tasks: tasks}
	wrapper, err := communication.Wrap(input)
	assert.NoError(t, err)
	b := testHelper.FakeTxCreator(t, org)

	stub := new(testHelper.MockedStub)
	stub.On("GetCreator").Return(b, nil).Once()

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("RegisterTasks", tasks, org).Return(nil).Once()

	ctx.On("GetStub").Return(stub).Once()

	_, err = contract.RegisterTasks(ctx, wrapper)
	assert.NoError(t, err, "task registration should not fail")
}
