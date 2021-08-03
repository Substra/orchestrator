package ledger

import (
	"context"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/peer"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
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

	stub.On("GetStateByPartialCompositeKey", computePlanTaskStatusIndex, []string{
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

func TestQueryComputePlans(t *testing.T) {
	stub := new(testHelper.MockedStub)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub)

	resp := &testHelper.MockedStateQueryIterator{}
	resp.On("Close").Return(nil)
	resp.On("HasNext").Once().Return(true)
	resp.On("HasNext").Once().Return(false)
	resp.On("Next").Once().Return(&queryresult.KV{Value: []byte(`{"asset":{"key":"uuid"}}`)}, nil)

	stub.On("GetState", "computeplan:uuid").Once().Return([]byte(`{"asset":{"key":"uuid"}}`), nil)
	idxIterator := &testHelper.MockedStateQueryIterator{}
	idxIterator.On("Close").Return(nil)
	// 2 calls for general index
	idxIterator.On("HasNext").Once().Return(true)
	idxIterator.On("HasNext").Once().Return(false)
	// 2 calls for DONE index
	idxIterator.On("HasNext").Once().Return(true)
	idxIterator.On("HasNext").Once().Return(false)
	idxIterator.On("Next").Return(&queryresult.KV{Key: "computeplan~uuid~STATUS_DONE~task1"}, nil)
	stub.On("SplitCompositeKey", "computeplan~uuid~STATUS_DONE~task1").Return("", []string{"computeplan", "uuid", "STATUS_DONE", "task1"}, nil)

	stub.On("GetStateByPartialCompositeKey", "computePlan~computePlanKey~status~task", []string{"computeplan", "uuid"}).Once().Return(idxIterator, nil)
	stub.On("GetStateByPartialCompositeKey", "computePlan~computePlanKey~status~task", []string{"computeplan", "uuid", "STATUS_DONE"}).Once().Return(idxIterator, nil)

	queryString := `{"selector":{"doc_type":"computeplan"},"fields":["asset.key"]}`
	stub.On("GetQueryResultWithPagination", queryString, int32(1), "").
		Return(resp, &peer.QueryResponseMetadata{Bookmark: "", FetchedRecordsCount: 1}, nil)

	_, _, err := db.QueryComputePlans(common.NewPagination("", 1))
	assert.NoError(t, err)
}
