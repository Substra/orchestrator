package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestAlgoServerImplementServer(t *testing.T) {
	server := NewAlgoServer()
	assert.Implementsf(t, (*asset.AlgoServiceServer)(nil), server, "AlgoServer should implements AlgoServiceServer")
}
