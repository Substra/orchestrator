package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestDataSampleServerImplementServer(t *testing.T) {
	server := NewDataSampleServer()
	assert.Implementsf(t, (*asset.DataSampleServiceServer)(nil), server, "DataSampleServer should implements DataSampleServiceServer")
}
