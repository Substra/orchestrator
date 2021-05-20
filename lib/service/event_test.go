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

package service

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	dispatcher := new(MockDispatcher)
	provider := new(MockServiceProvider)

	provider.On("GetEventDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)

	service := NewEventService(provider)

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  "uuid",
	}

	dbal.On("AddEvent", event).Once().Return(nil)
	dispatcher.On("Enqueue", event).Once().Return(nil)

	err := service.RegisterEvent(event)
	assert.NoError(t, err)

	assert.NotEqual(t, "", event.Id, "RegisterEvent should assign an ID to the event")

	provider.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
	dbal.AssertExpectations(t)
}
