package persistence

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestGetPlanStatus(t *testing.T) {
	cases := map[string]struct {
		total    uint32
		done     uint32
		doing    uint32
		waiting  uint32
		failed   uint32
		canceled uint32
		outcome  asset.ComputePlanStatus
	}{
		"done": {
			total:    11,
			done:     11,
			doing:    0,
			waiting:  0,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_DONE,
		},
		"waiting": {
			total:    11,
			done:     0,
			doing:    0,
			waiting:  11,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_WAITING,
		},
		"failed": {
			total:    11,
			done:     1,
			doing:    0,
			waiting:  1,
			failed:   1,
			canceled: 1,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_FAILED,
		},
		"canceled": {
			total:    11,
			done:     1,
			doing:    0,
			waiting:  1,
			failed:   0,
			canceled: 1,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_CANCELED,
		},
		"doing": {
			total:    11,
			done:     1,
			doing:    0,
			waiting:  1,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_DOING,
		},
		"todo": {
			total:    11,
			done:     0,
			doing:    0,
			waiting:  10,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_TODO,
		},
		"empty": {
			total:    0,
			done:     0,
			doing:    0,
			waiting:  0,
			failed:   0,
			canceled: 0,
			outcome:  asset.ComputePlanStatus_PLAN_STATUS_UNKNOWN,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			count := ComputePlanTaskCount{
				Total:    tc.total,
				Done:     tc.done,
				Doing:    tc.doing,
				Waiting:  tc.waiting,
				Failed:   tc.failed,
				Canceled: tc.canceled,
			}

			status := count.GetPlanStatus()
			assert.Equal(t, tc.outcome, status)
		})
	}

}
