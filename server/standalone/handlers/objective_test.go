package handlers

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestMetricServerImplementServer(t *testing.T) {
	server := NewMetricServer()
	assert.Implementsf(t, (*asset.MetricServiceServer)(nil), server, "MetricServer should implements MetricServiceServer")
}
