package service

import (
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	dispatcher := new(event.MockDispatcher)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()

	provider.On("GetEventDBAL").Return(dbal)
	provider.On("GetEventQueue").Return(dispatcher)
	provider.On("GetTimeService").Return(ts)

	service := NewEventService(provider)

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  "uuid",
	}

	dbal.On("AddEvents", event).Once().Return(nil)
	dispatcher.On("Enqueue", event).Once().Return(nil)
	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	err := service.RegisterEvents(event)
	assert.NoError(t, err)

	assert.NotEqual(t, "", event.Id, "RegisterEvents should assign an ID to the event")

	provider.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
	dbal.AssertExpectations(t)
}
