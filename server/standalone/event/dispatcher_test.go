// Package event contains AMQP dispatcher.
package event

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventChannel(t *testing.T) {
	amqp := new(common.MockPublisher)
	dispatcher := NewAMQPDispatcher(amqp, "testChannel")

	e := &asset.Event{AssetKind: asset.AssetKind_ASSET_NODE, AssetKey: "test", EventKind: asset.EventKind_EVENT_ASSET_CREATED}
	err := dispatcher.Enqueue(e)
	require.NoError(t, err)

	// Channel should be set on dispatch
	data := []byte(`{"id":"","asset_key":"test","asset_kind":"ASSET_NODE","event_kind":"EVENT_ASSET_CREATED","channel":"testChannel","timestamp":null,"metadata":{}}`)

	amqp.On("Publish", utils.AnyContext, "testChannel", data).Once().Return(nil)

	err = dispatcher.Dispatch(context.Background())
	assert.NoError(t, err)

	amqp.AssertExpectations(t)
}
