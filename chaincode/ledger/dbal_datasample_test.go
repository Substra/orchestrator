package ledger

import (
	"context"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/stretchr/testify/assert"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
)

func TestGetDataSampleKeysByManager(t *testing.T) {
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub, queue)

	resp := &testHelper.MockedStateQueryIterator{}
	resp.On("Close").Return(nil)
	resp.On("HasNext").Once().Return(false)
	resp.On("Next").Once().Return(&queryresult.KV{}, nil)

	queryString := `{"selector":{"doc_type":"datasample","asset":{"data_manager_keys":{"$elemMatch":{"$eq":"dmkey"}},"test_only":false}},"fields":["asset.key"]}`
	stub.On("GetQueryResult", queryString).Return(resp, nil)

	_, err := db.GetDataSampleKeysByManager("dmkey", false)
	assert.NoError(t, err)
}
