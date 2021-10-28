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

func TestQueryTaskEvents(t *testing.T) {
	stub := new(testHelper.MockedStub)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub)
	stub.On("GetChannelID").Return("eventTestChannel")

	iter := &testHelper.MockedStateQueryIterator{}
	iter.On("Close").Return(nil)
	iter.On("HasNext").Once().Return(false)
	iter.On("Next").Once().Return(&queryresult.KV{}, nil)

	queryString := `{"selector":{"doc_type":"event","asset":{"asset_key":"uuid","asset_kind":"ASSET_COMPUTE_TASK"}}}`
	stub.On("GetQueryResultWithPagination", queryString, int32(10), "").
		Return(iter, &peer.QueryResponseMetadata{Bookmark: "", FetchedRecordsCount: 1}, nil)

	pagination := common.NewPagination("", 10)

	filter := &asset.EventQueryFilter{
		AssetKey:  "uuid",
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
	}

	resp, _, err := db.QueryEvents(pagination, filter)
	assert.NoError(t, err)

	for _, event := range resp {
		assert.Equal(t, "eventTestChannel", event.Channel)
	}
}
