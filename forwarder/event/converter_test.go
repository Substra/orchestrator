package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/utils"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestForwardCCEvent(t *testing.T) {
	marshaller := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}
	events := []*asset.Event{
		{
			Id:        "event1",
			AssetKey:  "uuid1",
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			EventKind: asset.EventKind_EVENT_ASSET_CREATED,
			Timestamp: timestamppb.New(time.Unix(12, 0)),
		},
		{
			Id:        "event2",
			AssetKey:  "uuid1",
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
			Metadata:  map[string]string{"test": "value"},
			Timestamp: timestamppb.New(time.Unix(12, 0)),
		},
	}

	// Published event should have the channel set
	publishedEvents := []asset.Event{
		{
			Id:        "event1",
			AssetKey:  "uuid1",
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			EventKind: asset.EventKind_EVENT_ASSET_CREATED,
			Channel:   "testChannel",
			Timestamp: timestamppb.New(time.Unix(12, 0)),
		},
		{
			Id:        "event2",
			AssetKey:  "uuid1",
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
			Metadata:  map[string]string{"test": "value"},
			Channel:   "testChannel",
			Timestamp: timestamppb.New(time.Unix(12, 0)),
		},
	}

	ccEvents := make([]json.RawMessage, len(events))
	for i, event := range events {
		b, err := marshaller.Marshal(event)
		require.NoError(t, err)

		ccEvents[i] = b
	}

	payload, err := json.Marshal(ccEvents)
	require.NoError(t, err)

	ccEvent := &fab.CCEvent{Payload: payload}

	publisher := new(common.MockPublisher)
	forwarder := NewForwarder("testChannel", publisher)

	bytes1, err := marshaller.Marshal(&publishedEvents[0])
	require.NoError(t, err)
	bytes2, err := marshaller.Marshal(&publishedEvents[1])
	require.NoError(t, err)

	publisher.On("Publish", utils.AnyContext, "testChannel", bytes1).Once().Return(nil)
	publisher.On("Publish", utils.AnyContext, "testChannel", bytes2).Once().Return(nil)

	forwarder.Forward(ccEvent)

	publisher.AssertExpectations(t)
}
