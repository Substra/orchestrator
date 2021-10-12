package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerformanceGetKey(t *testing.T) {
	perf := &Performance{
		ComputeTaskKey: "taskKey",
		MetricKey:      "metricKey",
	}
	assert.Equal(t, perf.GetKey(), "taskKey|metricKey")
}
