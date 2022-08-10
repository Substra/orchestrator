package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

func TestDataManagerServerImplementServer(t *testing.T) {
	server := NewDataManagerServer()
	assert.Implementsf(t, (*asset.DataManagerServiceServer)(nil), server, "DataManagerServer should implements DataManagerServiceServer")
}
