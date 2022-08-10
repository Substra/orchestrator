package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

// TestInfoAdapterImplementServer makes sure the chaincode-baked orchestration exposes the same server than standalone mode.
func TestInfoAdapterImplementServer(t *testing.T) {
	adapter := NewInfoAdapter()
	assert.Implementsf(t, (*asset.InfoServiceServer)(nil), adapter, "InfoAdapter should implements InfoServiceServer")
}
