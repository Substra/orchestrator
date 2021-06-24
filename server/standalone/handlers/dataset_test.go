package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestDatasetServerImplementServer(t *testing.T) {
	server := NewDatasetServer()
	assert.Implementsf(t, (*asset.DatasetServiceServer)(nil), server, "DatasetServer should implements DatasetServiceServer")
}
