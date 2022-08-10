package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

func TestDatasetServerImplementServer(t *testing.T) {
	server := NewDatasetServer()
	assert.Implementsf(t, (*asset.DatasetServiceServer)(nil), server, "DatasetServer should implements DatasetServiceServer")
}
