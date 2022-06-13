package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

// TestOrganizationServerImplementServer makes sure chaincode-baked and standalone orchestration are in sync
func TestOrganizationServerImplementServer(t *testing.T) {
	server := NewOrganizationServer()
	assert.Implementsf(t, (*asset.OrganizationServiceServer)(nil), server, "OrganizationServer should implements OrganizationServiceServer")
}
