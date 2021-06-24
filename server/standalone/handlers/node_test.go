package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

// TestNodeServerImplementServer makes sure chaincode-baked and standalone orchestration are in sync
func TestNodeServerImplementServer(t *testing.T) {
	server := NewNodeServer()
	assert.Implementsf(t, (*asset.NodeServiceServer)(nil), server, "NodeServer should implements NodeServiceServer")
}
