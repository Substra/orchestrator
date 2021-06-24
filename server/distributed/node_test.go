package distributed

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

// TestNodeAdapterImplementServer makes sure the chaincode-baked orchestration exposes the same server than standalone mode.
func TestNodeAdapterImplementServer(t *testing.T) {
	adapter := NewNodeAdapter()
	assert.Implementsf(t, (*asset.NodeServiceServer)(nil), adapter, "NodeAdapter should implements NodeServiceServer")
}
