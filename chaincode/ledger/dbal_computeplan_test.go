package ledger

import (
	"context"
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

			count := planTaskCount{
				total:    tc.total,
				done:     tc.done,
				doing:    tc.doing,
				waiting:  tc.waiting,
				failed:   tc.failed,
				canceled: tc.canceled,
			}

			status := count.getPlanStatus()
			assert.Equal(t, tc.outcome, status)
		})
	}

}

func TestQueryComputePlans(t *testing.T) {
	stub := new(testHelper.MockedStub)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub)

	// ComputePlan iterator
	resp := &testHelper.MockedStateQueryIterator{}
	resp.On("Close").Return(nil)
	resp.On("HasNext").Once().Return(true)
	resp.On("HasNext").Once().Return(false)
	resp.On("Next").Once().Return(&queryresult.KV{Value: []byte(`{"asset":{"key":"uuid"}}`)}, nil)

	queryString := `{"selector":{"doc_type":"computeplan"}}`
	stub.On("GetQueryResultWithPagination", queryString, int32(1), "").
		Return(resp, &peer.QueryResponseMetadata{Bookmark: "", FetchedRecordsCount: 1}, nil)

	// CP tasks index
	index := &testHelper.MockedStateQueryIterator{}
	index.On("Close").Return(nil)
	index.On("HasNext").Twice().Return(true)
	index.On("HasNext").Once().Return(false)
	index.On("Next").Return(&queryresult.KV{Key: "firstIndexKey"}, nil)

	stub.On("SplitCompositeKey", "firstIndexKey").Return("", []string{"indexName", "uuid", "STATUS_FAILED", "taskId"}, nil)

	stub.On("GetStateByPartialCompositeKey", computePlanTaskStatusIndex, []string{asset.ComputePlanKind, "uuid"}).
		Return(index, nil)

	_, _, err := db.QueryComputePlans(common.NewPagination("", 1))
	assert.NoError(t, err)

	stub.AssertExpectations(t)
}
