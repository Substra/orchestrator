package model

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	"github.com/owkin/orchestrator/chaincode/mocks"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *mocks.TransactionContext) *service.MockModelAPI {
	mockService := new(service.MockModelAPI)

	provider := new(service.MockDependenciesProvider)
	provider.On("GetModelService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()
	ctx.On("SetRequestID", "").Once()
	ctx.On("GetContext").Return(context.Background())

	return mockService
}

func TestGetTaskOutputModels(t *testing.T) {
	contract := &SmartContract{}

	param := &asset.GetComputeTaskModelsParam{
		ComputeTaskKey: "uuid",
	}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	ctx := new(mocks.TransactionContext)

	service := getMockedService(ctx)
	service.On("GetComputeTaskOutputModels", "uuid").Return([]*asset.Model{{}, {}}, nil).Once()

	wrapped, err := contract.GetComputeTaskOutputModels(ctx, wrapper)
	assert.NoError(t, err)

	resp := new(asset.GetComputeTaskModelsResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)

	assert.Len(t, resp.Models, 2)
}

func TestGetTaskInputModels(t *testing.T) {
	contract := &SmartContract{}

	param := &asset.GetComputeTaskModelsParam{
		ComputeTaskKey: "uuid",
	}
	wrapper, err := communication.Wrap(context.Background(), param)
	assert.NoError(t, err)

	ctx := new(mocks.TransactionContext)

	service := getMockedService(ctx)
	service.On("GetComputeTaskInputModels", "uuid").Return([]*asset.Model{{}, {}}, nil).Once()

	wrapped, err := contract.GetComputeTaskInputModels(ctx, wrapper)
	assert.NoError(t, err)

	resp := new(asset.GetComputeTaskModelsResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)

	assert.Len(t, resp.Models, 2)
}

func TestRegisterModel(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"

	newModel := &asset.NewModel{
		Key:            "uuid",
		Category:       asset.ModelCategory_MODEL_SIMPLE,
		ComputeTaskKey: "taskUuid",
		Address:        &asset.Addressable{},
	}
	wrapper, err := communication.Wrap(context.Background(), newModel)
	assert.NoError(t, err)

	model := &asset.Model{}

	ctx := new(mocks.TransactionContext)

	service := getMockedService(ctx)
	service.On("RegisterModel", newModel, mspid).Return(model, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	_, err = contract.RegisterModel(ctx, wrapper)
	assert.NoError(t, err)
}

func TestCanDisableModel(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"

	wrapper, err := communication.Wrap(context.Background(), &asset.CanDisableModelParam{ModelKey: "uuid"})
	assert.NoError(t, err)

	ctx := new(mocks.TransactionContext)

	service := getMockedService(ctx)
	service.On("CanDisableModel", "uuid", mspid).Return(true, nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	wrapped, err := contract.CanDisableModel(ctx, wrapper)
	assert.NoError(t, err)

	resp := new(asset.CanDisableModelResponse)
	err = wrapped.Unwrap(resp)
	assert.NoError(t, err)

	assert.True(t, resp.CanDisable)
}

func TestDisableModel(t *testing.T) {
	contract := &SmartContract{}

	mspid := "org"

	wrapper, err := communication.Wrap(context.Background(), &asset.CanDisableModelParam{ModelKey: "uuid"})
	assert.NoError(t, err)

	ctx := new(mocks.TransactionContext)

	service := getMockedService(ctx)
	service.On("DisableModel", "uuid", mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.DisableModel(ctx, wrapper)
	assert.NoError(t, err)
}
