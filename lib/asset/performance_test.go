package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerformanceGetKey(t *testing.T) {
	perf := &Performance{
		ComputeTaskKey:              "taskKey",
		MetricKey:                   "metricKey",
		ComputeTaskOutputIdentifier: "performance",
	}
	assert.Equal(t, perf.GetKey(), "taskKey|metricKey|performance")
}
