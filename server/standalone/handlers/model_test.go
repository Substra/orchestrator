package handlers

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/standalone/interceptors"

	"github.com/stretchr/testify/assert"
)

func getContext() (context.Context, *service.MockServiceProvider) {
	provider := new(service.MockServiceProvider)
	ctx := context.TODO()
	ctxWithProvider := interceptors.WithProvider(ctx, provider)
	ctxWithIdentity := context.WithValue(ctxWithProvider, common.CtxMSPIDKey, "requester")

	return ctxWithIdentity, provider
}

func TestModelServiceServer(t *testing.T) {
	server := NewModelServer()
	assert.Implements(t, (*asset.ModelServiceServer)(nil), server)
}

func TestRegisterModel(t *testing.T) {
	ctx, p := getContext()
	ms := new(service.MockModelService)

	server := NewModelServer()

	newModel := &asset.NewModel{Key: "uuid"}

	p.On("GetModelService").Return(ms)
	ms.On("RegisterModel", newModel, "requester").Once().Return(&asset.Model{Key: "uuid"}, nil)

	_, err := server.RegisterModel(ctx, newModel)
	assert.NoError(t, err)

	p.AssertExpectations(t)
	ms.AssertExpectations(t)
}

func TestGetComputeTaskOutputModels(t *testing.T) {
	ctx, p := getContext()
	ms := new(service.MockModelService)

	server := NewModelServer()

	p.On("GetModelService").Return(ms)
	ms.On("GetComputeTaskOutputModels", "uuid").Once().Return([]*asset.Model{{Key: "m1"}, {Key: "m2"}}, nil)

	resp, err := server.GetComputeTaskOutputModels(ctx, &asset.GetComputeTaskModelsParam{ComputeTaskKey: "uuid"})
	assert.NoError(t, err)

	assert.Len(t, resp.Models, 2)

	p.AssertExpectations(t)
	ms.AssertExpectations(t)
}

func TestGetComputeTaskInputModels(t *testing.T) {
	ctx, p := getContext()
	ms := new(service.MockModelService)

	server := NewModelServer()

	p.On("GetModelService").Return(ms)
	ms.On("GetComputeTaskInputModels", "uuid").Once().Return([]*asset.Model{{Key: "m1"}, {Key: "m2"}}, nil)

	resp, err := server.GetComputeTaskInputModels(ctx, &asset.GetComputeTaskModelsParam{ComputeTaskKey: "uuid"})
	assert.NoError(t, err)

	assert.Len(t, resp.Models, 2)

	p.AssertExpectations(t)
	ms.AssertExpectations(t)
}

func TestCanDisableModel(t *testing.T) {
	ctx, p := getContext()
	ms := new(service.MockModelService)

	server := NewModelServer()

	p.On("GetModelService").Return(ms)
	ms.On("CanDisableModel", "uuid", "requester").Once().Return(true, nil)

	resp, err := server.CanDisableModel(ctx, &asset.CanDisableModelParam{ModelKey: "uuid"})
	assert.NoError(t, err)
	assert.True(t, resp.CanDisable)

	p.AssertExpectations(t)
	ms.AssertExpectations(t)
}

func TestDisableModel(t *testing.T) {
	ctx, p := getContext()
	ms := new(service.MockModelService)

	server := NewModelServer()

	p.On("GetModelService").Return(ms)
	ms.On("DisableModel", "uuid", "requester").Once().Return(nil)

	_, err := server.DisableModel(ctx, &asset.DisableModelParam{ModelKey: "uuid"})
	assert.NoError(t, err)

	p.AssertExpectations(t)
	ms.AssertExpectations(t)
}
