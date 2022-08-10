// Package event contains AMQP dispatcher.
package event

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/server/common"
	"github.com/substra/orchestrator/utils"
)

func TestEventChannel(t *testing.T) {
	amqp := new(common.MockAMQPPublisher)
	dispatcher := NewAMQPDispatcher(amqp, "testChannel")

	e := &asset.Event{AssetKind: asset.AssetKind_ASSET_ORGANIZATION, AssetKey: "test", EventKind: asset.EventKind_EVENT_ASSET_CREATED}
	err := dispatcher.Enqueue(e)
	require.NoError(t, err)

	// Channel should be set on dispatch
	dispatched := &asset.Event{AssetKind: asset.AssetKind_ASSET_ORGANIZATION, AssetKey: "test", EventKind: asset.EventKind_EVENT_ASSET_CREATED, Channel: "testChannel"}
	data, err := dispatcher.marshaller.Marshal(dispatched)
	require.NoError(t, err)

	amqp.On("Publish", utils.AnyContext, "testChannel", [][]byte{data}).Once().Return(nil)

	err = dispatcher.Dispatch(context.Background())
	assert.NoError(t, err)

	amqp.AssertExpectations(t)
}
