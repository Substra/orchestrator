package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestDataManagerServerImplementServer(t *testing.T) {
	server := NewDataManagerServer()
	assert.Implementsf(t, (*asset.DataManagerServiceServer)(nil), server, "DataManagerServer should implements DataManagerServiceServer")
}
