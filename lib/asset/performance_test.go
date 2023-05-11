package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerformanceGetKey(t *testing.T) {
	perf := &Performance{
		ComputeTaskKey:              "taskKey",
		ComputeTaskOutputIdentifier: "performance",
	}
	assert.Equal(t, perf.GetKey(), "taskKey|performance")
}
