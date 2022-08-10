package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

func TestAlgoServerImplementServer(t *testing.T) {
	server := NewAlgoServer()
	assert.Implementsf(t, (*asset.AlgoServiceServer)(nil), server, "AlgoServer should implements AlgoServiceServer")
}
