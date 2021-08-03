package service

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/event"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	dispatcher := new(event.MockDispatcher)
	provider := newMockedProvider()

	provider.On("GetEventDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)

	service := NewEventService(provider)

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  "uuid",
	}

	dbal.On("AddEvents", event).Once().Return(nil)
	dispatcher.On("Enqueue", event).Once().Return(nil)

	err := service.RegisterEvents(event)
	assert.NoError(t, err)

	assert.NotEqual(t, "", event.Id, "RegisterEvents should assign an ID to the event")

	provider.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
	dbal.AssertExpectations(t)
}
