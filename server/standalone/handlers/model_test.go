package handlers

import (
	"context"
	"testing"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/service"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/interceptors"

	"github.com/stretchr/testify/assert"
)

func getContext() (context.Context, *service.MockDependenciesProvider) {
	provider := new(service.MockDependenciesProvider)
	ctx := context.TODO()
	ctxWithProvider := interceptors.WithProvider(ctx, provider)
	ctxWithIdentity := context.WithValue(ctxWithProvider, commonInterceptors.CtxMSPIDKey, "requester")

	return ctxWithIdentity, provider
}

func TestModelServiceServer(t *testing.T) {
	server := NewModelServer()
	assert.Implements(t, (*asset.ModelServiceServer)(nil), server)
}

func TestRegisterModel(t *testing.T) {
	ctx, p := getContext()
	ms := new(service.MockModelAPI)

	server := NewModelServer()

	newModel := &asset.NewModel{Key: "uuid"}

	p.On("GetModelService").Return(ms)
	ms.On("RegisterModels", []*asset.NewModel{newModel}, "requester").Once().Return([]*asset.Model{{Key: "uuid"}}, nil)

	_, err := server.RegisterModel(ctx, newModel)
	assert.NoError(t, err)

	p.AssertExpectations(t)
	ms.AssertExpectations(t)
}

func TestGetComputeTaskOutputModels(t *testing.T) {
	ctx, p := getContext()
	ms := new(service.MockModelAPI)

	server := NewModelServer()

	p.On("GetModelService").Return(ms)
	ms.On("GetComputeTaskOutputModels", "uuid").Once().Return([]*asset.Model{{Key: "m1"}, {Key: "m2"}}, nil)

	resp, err := server.GetComputeTaskOutputModels(ctx, &asset.GetComputeTaskModelsParam{ComputeTaskKey: "uuid"})
	assert.NoError(t, err)

	assert.Len(t, resp.Models, 2)

	p.AssertExpectations(t)
	ms.AssertExpectations(t)
}
