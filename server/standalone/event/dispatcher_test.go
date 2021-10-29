// Package event contains AMQP dispatcher.
package event

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
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
	eventWithChannel := &asset.Event{AssetKind: asset.AssetKind_ASSET_NODE, AssetKey: "test", EventKind: asset.EventKind_EVENT_ASSET_CREATED, Channel: "testChannel"}

	data, err := json.Marshal(eventWithChannel)
	require.NoError(t, err)

	amqp.On("Publish", context.TODO(), "testChannel", data).Once().Return(nil)

	err = dispatcher.Dispatch(context.Background())
	assert.NoError(t, err)
}
