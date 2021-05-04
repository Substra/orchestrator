// Copyright 2020 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"testing"

	"github.com/owkin/orchestrator/chaincode/communication"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

// getMockedService returns a service mocks and make sure the provider returns the mock as well.
func getMockedService(ctx *testHelper.MockedContext) *service.MockModelService {
	mockService := new(service.MockModelService)

	provider := new(service.MockServiceProvider)
	provider.On("GetModelService").Return(mockService).Once()

	ctx.On("GetProvider").Return(provider).Once()

	return mockService
}

func TestGetTaskOutputModels(t *testing.T) {
	contract := &SmartContract{}

	param := &asset.GetComputeTaskModelsParam{
		ComputeTaskKey: "uuid",
	}
	wrapper, err := communication.Wrap(param)
	assert.NoError(t, err)

	ctx := new(testHelper.MockedContext)

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
	wrapper, err := communication.Wrap(param)
	assert.NoError(t, err)

	ctx := new(testHelper.MockedContext)

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
	wrapper, err := communication.Wrap(newModel)
	assert.NoError(t, err)

	model := &asset.Model{}

	ctx := new(testHelper.MockedContext)

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

	wrapper, err := communication.Wrap(&asset.CanDisableModelParam{ModelKey: "uuid"})
	assert.NoError(t, err)

	ctx := new(testHelper.MockedContext)

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

	wrapper, err := communication.Wrap(&asset.CanDisableModelParam{ModelKey: "uuid"})
	assert.NoError(t, err)

	ctx := new(testHelper.MockedContext)

	service := getMockedService(ctx)
	service.On("DisableModel", "uuid", mspid).Return(nil).Once()

	stub := new(testHelper.MockedStub)
	ctx.On("GetStub").Return(stub).Once()

	stub.On("GetCreator").Return(testHelper.FakeTxCreator(t, mspid), nil).Once()

	err = contract.DisableModel(ctx, wrapper)
	assert.NoError(t, err)
}
