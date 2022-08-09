package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
)

func TestComputeTaskServerImplementServer(t *testing.T) {
	server := NewComputeTaskServer()
	assert.Implements(t, (*asset.ComputeTaskServiceServer)(nil), server)
}

func TestGetTaskInputAssets(t *testing.T) {
	ctx, p := getContext()
	cts := new(service.MockComputeTaskAPI)

	server := NewComputeTaskServer()

	inputs := []*asset.ComputeTaskInputAsset{
		{Identifier: "test"},
	}

	p.On("GetComputeTaskService").Return(cts)
	cts.On("GetInputAssets", "uuid").Once().Return(inputs, nil)

	resp, err := server.GetTaskInputAssets(ctx, &asset.GetTaskInputAssetsParam{ComputeTaskKey: "uuid"})
	assert.NoError(t, err)
	assert.Equal(t, inputs, resp.Assets)

	p.AssertExpectations(t)
	cts.AssertExpectations(t)
}
