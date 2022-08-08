package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type performanceTestCase struct {
	performance *NewPerformance
	valid       bool
}

func TestPerformanceValidate(t *testing.T) {
	cases := map[string]performanceTestCase{
		"emtpy": {&NewPerformance{}, false},
		"invalidComputeTaskKey": {&NewPerformance{
			ComputeTaskKey:              "not36chars",
			ComputeTaskOutputIdentifier: "auc",
			MetricKey:                   "1da600d4-f8ad-45d7-92a0-7ff752a82275",
			PerformanceValue:            0.5,
		}, false},
		"invalidMetricKey": {&NewPerformance{
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "auc",
			MetricKey:                   "not36chars",
			PerformanceValue:            0.5,
		}, false},
		"missingComputeTaskOutput": {&NewPerformance{
			ComputeTaskKey:   "08680966-97ae-4573-8b2d-6c4db2b3c532",
			MetricKey:        "not36chars",
			PerformanceValue: 0.5,
		}, false},
		"valid": {&NewPerformance{
			ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
			ComputeTaskOutputIdentifier: "auc",
			MetricKey:                   "1da600d4-f8ad-45d7-92a0-7ff752a82275",
			PerformanceValue:            0.5,
		}, true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tc.valid {
				assert.NoError(t, tc.performance.Validate())
			} else {
				assert.Error(t, tc.performance.Validate())
			}
		})
	}
}
