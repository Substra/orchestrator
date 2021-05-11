// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ledger

import (
	"context"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestGetPlanStatus(t *testing.T) {
	cases := map[string]struct {
		total    int
		done     int
		doing    int
		waiting  int
		failed   int
		canceled int
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
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			stub := new(testHelper.MockedStub)

			mockIndexCount(stub, asset.ComputeTaskStatus_STATUS_FAILED, tc.failed)
			mockIndexCount(stub, asset.ComputeTaskStatus_STATUS_CANCELED, tc.canceled)
			mockIndexCount(stub, asset.ComputeTaskStatus_STATUS_WAITING, tc.waiting)
			mockIndexCount(stub, asset.ComputeTaskStatus_STATUS_DOING, tc.doing)

			db := NewDB(context.TODO(), stub)

			status, err := db.getPlanStatus("uuid", tc.total, tc.done)
			assert.NoError(t, err)
			assert.Equal(t, tc.outcome, status)
		})
	}

}

func mockIndexCount(stub *testHelper.MockedStub, status asset.ComputeTaskStatus, count int) {
	// should be unique for each status, but has no real impact on the test
	compKey := fmt.Sprintf("compkey+%s", status.String())

	stub.On("GetStateByPartialCompositeKey", indexPlanTaskStatus, []string{
		asset.ComputePlanKind, "uuid", status.String(),
	}).Return(getIterator(count, compKey), nil)

	stub.On("SplitCompositeKey", compKey).Return("", []string{"key"}, nil)
}

func getIterator(count int, key string) *testHelper.MockedStateQueryIterator {
	iter := &testHelper.MockedStateQueryIterator{}
	iter.On("Close").Return(nil)
	if count > 0 {
		iter.On("HasNext").Times(count).Return(true)
	}
	iter.On("HasNext").Once().Return(false)
	iter.On("Next").Return(&queryresult.KV{
		Key: key,
	}, nil)

	return iter
}
