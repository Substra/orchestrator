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

package service

import (
	"errors"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	eventtesting "github.com/owkin/orchestrator/lib/event/testing"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterNode(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)
	dispatcher := new(MockDispatcher)

	provider.On("GetNodeDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)

	e := &event.Event{EventKind: event.AssetCreated, AssetKind: asset.NodeKind, AssetKey: "uuid1"}
	dispatcher.On("Enqueue", mock.MatchedBy(eventtesting.EventMatcher(e))).Once().Return(nil)

	expected := asset.Node{
		Id: "uuid1",
	}

	dbal.On("NodeExists", "uuid1").Return(false, nil).Once()
	dbal.On("AddNode", &expected).Return(nil).Once()

	service := NewNodeService(provider)

	node, err := service.RegisterNode("uuid1")
	assert.NoError(t, err, "Node registration should not fail")
	assert.Equal(t, &expected, node, "Registration should return a node")
}

func TestRegisterExistingNode(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	provider := new(MockServiceProvider)

	provider.On("GetNodeDBAL").Return(dbal)

	dbal.On("NodeExists", "uuid1").Return(true, nil).Once()

	service := NewNodeService(provider)

	_, err := service.RegisterNode("uuid1")
	assert.Error(t, err, "Registration should fail for existing node")
	assert.True(t, errors.Is(err, orcerrors.ErrConflict))
}
