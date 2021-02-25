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

package orchestration

import (
	"errors"
	"testing"

	"github.com/owkin/orchestrator/lib/assets"
	orchestrationError "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterNode(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	provider := new(MockServiceProvider)
	dispatcher := new(MockDispatcher)

	provider.On("GetDatabase").Return(mockDB)
	provider.On("GetEventQueue").Return(dispatcher)

	e := &event.Event{EventKind: event.AssetCreated, AssetKind: assets.NodeKind, AssetID: "uuid1"}
	dispatcher.On("Enqueue", e).Return(nil)

	expected := assets.Node{
		Id: "uuid1",
	}

	mockDB.On("HasKey", assets.NodeKind, "uuid1", mock.Anything).Return(false, nil).Once()
	mockDB.On("PutState", assets.NodeKind, "uuid1", mock.Anything).Return(nil).Once()

	service := NewNodeService(provider)

	node, err := service.RegisterNode("uuid1")
	assert.NoError(t, err, "Node registration should not fail")
	assert.Equal(t, &expected, node, "Registration should return a node")
}

func TestRegisterExistingNode(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	provider := new(MockServiceProvider)

	provider.On("GetDatabase").Return(mockDB)

	mockDB.On("HasKey", assets.NodeKind, "uuid1", mock.Anything).Return(true, nil).Once()

	service := NewNodeService(provider)

	_, err := service.RegisterNode("uuid1")
	assert.Error(t, err, "Registration should fail for existing node")
	assert.True(t, errors.Is(err, orchestrationError.ErrConflict))
}
