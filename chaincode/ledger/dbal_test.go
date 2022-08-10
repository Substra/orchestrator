package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/assert"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
)

func TestGetFullKey(t *testing.T) {
	k := getFullKey("resource", "id")

	assert.Equal(t, "resource:id", k, "key should be prefixed with resource type")
}

// In this test we fake 2 metrics in db, and fetch them in two queries of pageSize 1
func TestGetPagination(t *testing.T) {
	stub := new(testHelper.MockedStub)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub)

	resp1 := &testHelper.MockedStateQueryIterator{}
	resp1.On("Close").Return(nil)
	resp1.On("HasNext").Once().Return(true)
	resp1.On("HasNext").Once().Return(false)
	resp1.On("Next").Once().Return(&queryresult.KV{}, nil)

	meta1 := &peer.QueryResponseMetadata{
		Bookmark:            "bmarkFromLedger1",
		FetchedRecordsCount: 1,
	}

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.MetricKind,
		},
	}

	b, err := json.Marshal(query)
	assert.NoError(t, err)
	queryString := string(b)
	//Notice how we request pagesize + 1 to check if we reached last page
	stub.On("GetQueryResultWithPagination", queryString, int32(1), "").Return(resp1, meta1, nil)

	_, firstBmark, err := db.getQueryResultWithPagination(queryString, int32(1), "")

	assert.NoError(t, err)
	assert.NotEqual(t, "", firstBmark, "bookmark should not be empty")

	resp2 := &testHelper.MockedStateQueryIterator{}
	resp2.On("Close").Return(nil)
	resp2.On("HasNext").Once().Return(true)
	resp2.On("HasNext").Once().Return(false)
	resp2.On("Next").Once().Return(&queryresult.KV{}, nil)

	meta2 := &peer.QueryResponseMetadata{
		Bookmark:            "bmarkFromLedger2",
		FetchedRecordsCount: 1, // here there is only 1 key left
	}

	// Notice how we request pagesize + 1 to check if we reached last page
	stub.On("GetQueryResultWithPagination", queryString, int32(1), firstBmark.Bookmark).Return(resp2, meta2, nil)
	_, _, err = db.getQueryResultWithPagination(queryString, 1, firstBmark.Bookmark)
	assert.NoError(t, err)
}

func TestValidateQueryContext(t *testing.T) {
	var db *DB
	var err error
	mockStub := new(testHelper.MockedStub)

	// no context: error
	db = NewDB(context.Background(), mockStub)
	err = db.validateQueryContext()
	orcErr := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcErr))
	assert.Equal(t, orcerrors.ErrInternal, orcErr.Kind)

	// context with isEval=false: error
	db = NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, false), mockStub)
	err = db.validateQueryContext()
	orcErr = new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcErr))
	assert.Equal(t, orcerrors.ErrInternal, orcErr.Kind)

	// context with isEval=true: ok
	db = NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), mockStub)
	err = db.validateQueryContext()
	assert.NoError(t, err)
}

// ensure CouchDB calls the validateQueryContext method()
func TestCheckQueryContext(t *testing.T) {
	var db *DB
	var err error
	mockStub := new(testHelper.MockedStub)

	// getQueryResult
	db = NewDB(context.Background(), mockStub)
	_, err = db.getQueryResult("some query")
	orcErr := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcErr))
	assert.Equal(t, orcerrors.ErrInternal, orcErr.Kind)

	// getQueryResultWithPagination
	db = NewDB(context.Background(), mockStub)
	_, _, err = db.getQueryResultWithPagination("some query", 0, "bookmark")
	orcErr = new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcErr))
	assert.Equal(t, orcerrors.ErrInternal, orcErr.Kind)
}

func TestTransactionState(t *testing.T) {
	mockStub := new(testHelper.MockedStub)

	db := NewDB(context.Background(), mockStub)
	fullkey := getFullKey("test", "key")
	storedAsset := &storedAsset{
		DocType: "test",
		Asset:   []byte("{}"),
	}

	sab, _ := json.Marshal(storedAsset)
	db.putTransactionState(fullkey, sab)

	b, err := db.getState("test", "key")
	assert.Equal(t, b, []byte("{}"))
	assert.NoError(t, err)
}

func TestRichQueryFilter(t *testing.T) {
	cases := map[string]struct {
		filter richQuerySelector
		output string
	}{
		"simple asset": {
			filter: richQuerySelector{
				Selector: couchAssetQuery{DocType: "event"},
			},
			output: `{"selector":{"doc_type":"event"}}`,
		},
		"fields": {
			filter: richQuerySelector{
				Selector: couchAssetQuery{DocType: "event"},
				Fields:   []string{"asset_key"},
			},
			output: `{"selector":{"doc_type":"event"},"fields":["asset_key"]}`,
		},
		"asset filter": {
			filter: richQuerySelector{
				Selector: couchAssetQuery{DocType: "event", Asset: map[string]interface{}{"asset_key": "uuid"}},
			},
			output: `{"selector":{"doc_type":"event","asset":{"asset_key":"uuid"}}}`,
		},
		"sort": {
			filter: richQuerySelector{
				Selector: couchAssetQuery{DocType: "event"},
				Sort:     []map[string]string{{"asset.timestamp": CouchDBSortAsc}},
			},
			output: `{"selector":{"doc_type":"event"},"sort":[{"asset.timestamp":"asc"}]}`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			b, err := json.Marshal(tc.filter)
			assert.NoError(t, err)
			assert.Equal(t, tc.output, string(b))
		})
	}
}
