// Copyright 2021 Owkin Inc.
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
