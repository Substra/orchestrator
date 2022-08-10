package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetPlanStatus(t *testing.T) {
	cases := map[string]struct {
		total           uint32
		done            uint32
		doing           uint32
		waiting         uint32
		failed          uint32
		canceled        uint32
		outcome         asset.ComputePlanStatus
		cancelationDate *timestamppb.Timestamp
	}{
		"done": {
			total:           11,
			done:            11,
			doing:           0,
			waiting:         0,
			failed:          0,
			canceled:        0,
			outcome:         asset.ComputePlanStatus_PLAN_STATUS_DONE,
			cancelationDate: nil,
		},
		"waiting": {
			total:           11,
			done:            0,
			doing:           0,
			waiting:         11,
			failed:          0,
			canceled:        0,
			outcome:         asset.ComputePlanStatus_PLAN_STATUS_WAITING,
			cancelationDate: nil,
		},
		"failed": {
			total:           11,
			done:            1,
			doing:           0,
			waiting:         1,
			failed:          1,
			canceled:        1,
			outcome:         asset.ComputePlanStatus_PLAN_STATUS_FAILED,
			cancelationDate: nil,
		},
		"canceled count": {
			total:           11,
			done:            1,
			doing:           0,
			waiting:         1,
			failed:          0,
			canceled:        1,
			outcome:         asset.ComputePlanStatus_PLAN_STATUS_CANCELED,
			cancelationDate: nil,
		},
		"canceled CancelationDate": {
			total:           11,
			done:            1,
			doing:           0,
			waiting:         1,
			failed:          0,
			canceled:        0,
			outcome:         asset.ComputePlanStatus_PLAN_STATUS_CANCELED,
			cancelationDate: timestamppb.Now(),
		},
		"doing": {
			total:           11,
			done:            1,
			doing:           0,
			waiting:         1,
			failed:          0,
			canceled:        0,
			outcome:         asset.ComputePlanStatus_PLAN_STATUS_DOING,
			cancelationDate: nil,
		},
		"todo": {
			total:           11,
			done:            0,
			doing:           0,
			waiting:         10,
			failed:          0,
			canceled:        0,
			outcome:         asset.ComputePlanStatus_PLAN_STATUS_TODO,
			cancelationDate: nil,
		},
		"empty": {
			total:           0,
			done:            0,
			doing:           0,
			waiting:         0,
			failed:          0,
			canceled:        0,
			outcome:         asset.ComputePlanStatus_PLAN_STATUS_EMPTY,
			cancelationDate: nil,
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

			status := GetPlanStatus(&asset.ComputePlan{CancelationDate: tc.cancelationDate}, &count)
			assert.Equal(t, tc.outcome, status)
		})
	}

}
