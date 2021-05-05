// Copyright 2020 Owkin Inc.
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
	"encoding/json"
	"errors"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/peer"
	testHelper "github.com/owkin/orchestrator/chaincode/testing"
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetFullKey(t *testing.T) {
	k := getFullKey("resource", "id")

	assert.Equal(t, "resource:id", k, "key should be prefixed with resource type")
}

// In this test we fake 2 objectives in db, and fetch them in two queries of pageSize 1
func TestGetPagination(t *testing.T) {
	stub := new(testHelper.MockedStub)
	stub.On("SplitCompositeKey", mock.Anything).Return("", []string{"1", "2", "2"}, nil) // Don't really care about keys here

	db := NewDB(context.TODO(), stub)

	index := "objective~owner~key"
	attributes := []string{"objective"}

	resp1 := &testHelper.MockedStateQueryIterator{}
	resp1.On("Close").Return(nil)
	resp1.On("HasNext").Once().Return(true)
	resp1.On("HasNext").Once().Return(false)
	resp1.On("Next").Once().Return(&queryresult.KV{}, nil)

	meta1 := &peer.QueryResponseMetadata{
		Bookmark:            "bmarkFromLedger1",
		FetchedRecordsCount: 1,
	}

	// Notice how we request pagesize + 1 to check if we reached last page
	stub.On("GetStateByPartialCompositeKeyWithPagination", index, attributes, int32(1), "").Return(resp1, meta1, nil)

	_, firstBmark, err := db.getIndexKeysWithPagination(index, attributes, 1, "")
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
	stub.On("GetStateByPartialCompositeKeyWithPagination", index, attributes, int32(1), firstBmark).Return(resp2, meta2, nil)

	_, _, err = db.getIndexKeysWithPagination(index, attributes, 1, firstBmark)
	assert.NoError(t, err)

}

func TestAddExistingObjective(t *testing.T) {
	stub := new(testHelper.MockedStub)

	db := NewDB(context.TODO(), stub)

	objective := &asset.Objective{Key: "test"}

	stub.On("GetState", "objective:test").Return([]byte("{}"), nil).Once()

	err := db.AddObjective(objective)
	assert.True(t, errors.Is(err, orcerrors.ErrConflict))
}

func TestValidateQueryContext(t *testing.T) {
	var db *DB
	var err error
	mockStub := new(testHelper.MockedStub)

	// no context: error
	db = NewDB(context.Background(), mockStub)
	err = db.validateQueryContext()
	assert.True(t, errors.Is(err, orcerrors.ErrInternalError))

	// context with isEval=false: error
	db = NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, false), mockStub)
	err = db.validateQueryContext()
	assert.True(t, errors.Is(err, orcerrors.ErrInternalError))

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
	assert.True(t, errors.Is(err, orcerrors.ErrInternalError))

	// getQueryResultWithPagination
	db = NewDB(context.Background(), mockStub)
	_, _, err = db.getQueryResultWithPagination("some query", 0, "bookmark")
	assert.True(t, errors.Is(err, orcerrors.ErrInternalError))
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
