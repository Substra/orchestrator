package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestObjectiveServerImplementServer(t *testing.T) {
	server := NewObjectiveServer()
	assert.Implementsf(t, (*asset.ObjectiveServiceServer)(nil), server, "ObjectiveServer should implements ObjectiveServiceServer")
}
