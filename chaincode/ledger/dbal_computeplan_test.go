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

func TestQueryComputePlans(t *testing.T) {
	stub := new(testHelper.MockedStub)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub)

	// ComputePlan iterator
	resp := &testHelper.MockedStateQueryIterator{}
	resp.On("Close").Return(nil)
	resp.On("HasNext").Once().Return(true)
	resp.On("HasNext").Once().Return(false)
	resp.On("Next").Once().Return(&queryresult.KV{Value: []byte(`{"asset":{"key":"uuid"}}`)}, nil)

	queryString := `{"selector":{"doc_type":"computeplan","asset":{"owner":"owner"}}}`
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

	_, _, err := db.QueryComputePlans(
		common.NewPagination("", 1),
		&asset.PlanQueryFilter{Owner: "owner"},
	)
	assert.NoError(t, err)

	stub.AssertExpectations(t)
}
