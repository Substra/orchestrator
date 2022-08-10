package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

func TestDataSampleServerImplementServer(t *testing.T) {
	server := NewDataSampleServer()
	assert.Implementsf(t, (*asset.DataSampleServiceServer)(nil), server, "DataSampleServer should implements DataSampleServiceServer")
}
