package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

func TestFunctionServerImplementServer(t *testing.T) {
	server := NewFunctionServer()
	assert.Implementsf(t, (*asset.FunctionServiceServer)(nil), server, "FunctionServer should implements FunctionServiceServer")
}
