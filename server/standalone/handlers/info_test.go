package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

func TestInfoServerImplementServer(t *testing.T) {
	server := NewInfoServer()
	assert.Implementsf(t, (*asset.InfoServiceServer)(nil), server, "InfoServer should implements InfoServiceServer")
}

func TestInfoServerReturnVersion(t *testing.T) {
	server := NewInfoServer()

	version, err := server.QueryVersion(context.TODO(), &asset.QueryVersionParam{})
	assert.Equal(t, version.Orchestrator, "dev", "Orchestrator version should match")
	assert.Equal(t, version.Chaincode, "", "Chaincode version should match")
	assert.NoError(t, err)
}
