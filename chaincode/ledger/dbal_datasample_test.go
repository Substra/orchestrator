package ledger

import (
	"context"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/stretchr/testify/assert"
)

func TestGetDataSampleKeysByDataManager(t *testing.T) {
	stub := new(testHelper.MockedStub)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub)

	resp := &testHelper.MockedStateQueryIterator{}
	resp.On("Close").Return(nil)
	resp.On("HasNext").Once().Return(false)
	resp.On("Next").Once().Return(&queryresult.KV{}, nil)

	queryString := `{"selector":{"doc_type":"datasample","asset":{"data_manager_keys":{"$elemMatch":{"$eq":"dmkey"}},"test_only":false}},"fields":["asset.key"]}`
	stub.On("GetQueryResult", queryString).Return(resp, nil)

	_, err := db.GetDataSamplesKeysByDataManager("dmkey", false)
	assert.NoError(t, err)
}
