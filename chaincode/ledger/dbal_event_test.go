package ledger

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	testHelper "github.com/substra/orchestrator/chaincode/testing"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAddEvents(t *testing.T) {
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub, queue)

	event := &asset.Event{Id: "0", Asset: &asset.Event_Algo{Algo: &asset.Algo{}}}

	queue.On("Enqueue", event).Once().Return(nil)
	var buff []byte
	stub.On("GetState", "event:0").Once().Return(buff, nil)
	stub.On("PutState", "event:0", mock.AnythingOfType("[]uint8")).Once().Return(nil)

	err := db.AddEvents(event)
	assert.NoError(t, err)
	stub.AssertExpectations(t)
	queue.AssertExpectations(t)
}

func TestQueryTaskEvents(t *testing.T) {
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub, queue)
	stub.On("GetChannelID").Return("eventTestChannel")

	iter := &testHelper.MockedStateQueryIterator{}
	iter.On("Close").Return(nil)
	iter.On("HasNext").Once().Return(false)
	iter.On("Next").Once().Return(&queryresult.KV{}, nil)

	queryString := `{"selector":{"doc_type":"event","asset":{"asset_key":"uuid","asset_kind":"ASSET_COMPUTE_TASK"}},"sort":[{"asset.timestamp":"asc"},{"asset.id":"asc"}]}`
	stub.On("GetQueryResultWithPagination", queryString, int32(10), "").
		Return(iter, &peer.QueryResponseMetadata{Bookmark: "", FetchedRecordsCount: 1}, nil)

	pagination := common.NewPagination("", 10)

	filter := &asset.EventQueryFilter{
		AssetKey:  "uuid",
		AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
	}

	resp, _, err := db.QueryEvents(pagination, filter, asset.SortOrder_ASCENDING)
	assert.NoError(t, err)

	for _, event := range resp {
		assert.Equal(t, "eventTestChannel", event.Channel)
	}
}

// TestProxyFields should fail if the proxy object is not updated after a new field is added to the asset.
func TestProxyFields(t *testing.T) {
	var publicEventFields, publicProxyFields int

	eventType := reflect.TypeOf(asset.Event{})
	eventFields := reflect.VisibleFields(eventType)
	for _, f := range eventFields {
		if f.IsExported() {
			publicEventFields++
		}
	}

	proxyType := reflect.TypeOf(storableEvent{})
	proxyFields := reflect.VisibleFields(proxyType)
	for _, f := range proxyFields {
		if f.IsExported() {
			publicProxyFields++
		}
	}

	assert.GreaterOrEqual(t, publicProxyFields, publicEventFields, "proxy should have at least as many fields than the asset it represents")
}

func TestProxyConversion(t *testing.T) {
	event := &asset.Event{
		Id:        "test",
		AssetKey:  "testAsset",
		AssetKind: asset.AssetKind_ASSET_ALGO,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Channel:   "testChannel",
		Timestamp: timestamppb.New(time.Unix(1337, 1234)),
		Asset:     &asset.Event_Algo{Algo: new(asset.Algo)},
		Metadata:  map[string]string{"test": "true"},
	}

	proxy, err := newStorableEvent(event)
	require.NoError(t, err)
	assert.Equal(t, int64(1337000001234), proxy.Timestamp, "proxy object should store TS as unix time")

	converted, err := proxy.newEvent()
	require.NoError(t, err)
	assert.Equal(t, event, converted)
}

func TestFailingProxyConversion(t *testing.T) {
	cases := map[string]storableEvent{
		"invalid event": {
			EventKind: "notanevent",
		},
		"invalid asset": {
			AssetKind: "notanasset",
		},
	}

	for name, stored := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := stored.newEvent()
			assert.Error(t, err)
			orcError := new(orcerrors.OrcError)
			assert.True(t, errors.As(err, &orcError))
			assert.Equal(t, orcerrors.ErrUnimplemented, orcError.Kind)
		})
	}

}

func TestEventAssetFilterBuilder(t *testing.T) {
	cases := map[string]struct {
		input  *asset.EventQueryFilter
		output map[string]interface{}
	}{
		"asset_key": {
			input:  &asset.EventQueryFilter{AssetKey: "test"},
			output: map[string]interface{}{"asset_key": "test"},
		},
		"asset_kind": {
			input:  &asset.EventQueryFilter{AssetKind: asset.AssetKind_ASSET_ALGO},
			output: map[string]interface{}{"asset_kind": "ASSET_ALGO"},
		},
		"event_kind": {
			input:  &asset.EventQueryFilter{EventKind: asset.EventKind_EVENT_ASSET_CREATED},
			output: map[string]interface{}{"event_kind": "EVENT_ASSET_CREATED"},
		},
		"metadata": {
			input:  &asset.EventQueryFilter{Metadata: map[string]string{"test": "true"}},
			output: map[string]interface{}{"metadata": map[string]string{"test": "true"}},
		},
		"start": {
			input:  &asset.EventQueryFilter{Start: timestamppb.New(time.Unix(1337, 1234))},
			output: map[string]interface{}{"timestamp": map[string]int64{"$gte": 1337000001234}},
		},
		"bound": {
			input: &asset.EventQueryFilter{
				AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
				Start:     timestamppb.New(time.Unix(1337, 1234)),
				End:       timestamppb.New(time.Unix(1845, 1234)),
			},
			output: map[string]interface{}{
				"asset_kind": "ASSET_COMPUTE_TASK",
				"timestamp":  map[string]int64{"$gte": 1337000001234, "$lte": 1845000001234},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.output, buildEventAssetFilter(tc.input))
		})
	}
}

func TestQueryEventsNilFilter(t *testing.T) {
	stub := new(testHelper.MockedStub)
	queue := new(MockEventQueue)
	db := NewDB(context.WithValue(context.Background(), ctxIsEvaluateTransaction, true), stub, queue)
	stub.On("GetChannelID").Return("eventTestChannel")

	iter := &testHelper.MockedStateQueryIterator{}
	iter.On("Close").Return(nil)
	iter.On("HasNext").Once().Return(false)
	iter.On("Next").Once().Return(&queryresult.KV{}, nil)

	queryString := `{"selector":{"doc_type":"event"},"sort":[{"asset.timestamp":"asc"},{"asset.id":"asc"}]}`
	stub.On("GetQueryResultWithPagination", queryString, int32(10), "").
		Return(iter, &peer.QueryResponseMetadata{Bookmark: "", FetchedRecordsCount: 0}, nil)

	pagination := common.NewPagination("", 10)

	_, _, err := db.QueryEvents(pagination, nil, asset.SortOrder_ASCENDING)
	assert.NoError(t, err)
}
