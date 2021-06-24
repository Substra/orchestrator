package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEvaluateMethod(t *testing.T) {
	cases := map[string]struct {
		method     string
		isEvaluate bool
	}{
		"register task": {
			"/orchestrator.ComputeTaskService/RegisterTask",
			false,
		},
		"query task": {
			"/orchestrator.ComputeTaskService/QueryTasks",
			true,
		},
		"unknown": {
			"/orchestrator.ComputeTaskService/Unknown",
			false,
		},
		"bad format": {
			"/Unknown",
			false,
		},
	}

	grpcChecker := new(GrpcMethodChecker)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.isEvaluate, grpcChecker.IsEvaluateMethod(tc.method))
		})
	}
}
