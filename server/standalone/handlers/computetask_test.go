package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestComputeTaskServerImplementServer(t *testing.T) {
	server := NewComputeTaskServer()
	assert.Implements(t, (*asset.ComputeTaskServiceServer)(nil), server)
}
