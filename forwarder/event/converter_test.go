package event

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/stretchr/testify/require"
)

func TestForwardCCEvent(t *testing.T) {
	events := []asset.Event{
		{
			Id:        "event1",
			AssetKey:  "uuid1",
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			EventKind: asset.EventKind_EVENT_ASSET_CREATED,
			Timestamp: 12,
		},
		{
			Id:        "event2",
			AssetKey:  "uuid1",
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
			Metadata:  map[string]string{"test": "value"},
			Timestamp: 12,
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
			Timestamp: 12,
		},
		{
			Id:        "event2",
			AssetKey:  "uuid1",
			AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK,
			EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
			Metadata:  map[string]string{"test": "value"},
			Channel:   "testChannel",
			Timestamp: 12,
		},
	}

	payload, err := json.Marshal(events)
	require.NoError(t, err)

	ccEvent := &fab.CCEvent{Payload: payload}

	publisher := new(common.MockPublisher)
	forwarder := NewForwarder("testChannel", publisher)

	bytes1, err := json.Marshal(&publishedEvents[0])
	require.NoError(t, err)
	bytes2, err := json.Marshal(&publishedEvents[1])
	require.NoError(t, err)

	publisher.On("Publish", "testChannel", bytes1).Once().Return(nil)
	publisher.On("Publish", "testChannel", bytes2).Once().Return(nil)

	forwarder.Forward(ccEvent)
}
