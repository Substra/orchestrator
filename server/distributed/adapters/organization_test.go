package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

// TestOrganizationAdapterImplementServer makes sure the chaincode-baked orchestration exposes the same server than standalone mode.
func TestOrganizationAdapterImplementServer(t *testing.T) {
	adapter := NewOrganizationAdapter()
	assert.Implementsf(t, (*asset.OrganizationServiceServer)(nil), adapter, "OrganizationAdapter should implements OrganizationServiceServer")
}
