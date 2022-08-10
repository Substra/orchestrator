package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

// TestOrganizationServerImplementServer makes sure chaincode-baked and standalone orchestration are in sync
func TestOrganizationServerImplementServer(t *testing.T) {
	server := NewOrganizationServer()
	assert.Implementsf(t, (*asset.OrganizationServiceServer)(nil), server, "OrganizationServer should implements OrganizationServiceServer")
}
