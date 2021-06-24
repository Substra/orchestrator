// Package event contains AMQP dispatcher.
package event

import (
	"encoding/json"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAMQPChannel struct {
	mock.Mock
}

func (m *MockAMQPChannel) Publish(routingKey string, data []byte) error {
	args := m.Called(routingKey, data)
	return args.Error(0)
}

func TestEventChannel(t *testing.T) {
	amqp := &MockAMQPChannel{}
	dispatcher := NewAMQPDispatcher(amqp, "testChannel")

	e := &asset.Event{AssetKind: asset.AssetKind_ASSET_NODE, AssetKey: "test", EventKind: asset.EventKind_EVENT_ASSET_CREATED}
	err := dispatcher.Enqueue(e)
	require.NoError(t, err)

	// Channel should be set on dispatch
	eventWithChannel := &asset.Event{AssetKind: asset.AssetKind_ASSET_NODE, AssetKey: "test", EventKind: asset.EventKind_EVENT_ASSET_CREATED, Channel: "testChannel"}

	data, err := json.Marshal(eventWithChannel)
	require.NoError(t, err)

	amqp.On("Publish", "testChannel", data).Once().Return(nil)

	err = dispatcher.Dispatch()
	assert.NoError(t, err)
}
