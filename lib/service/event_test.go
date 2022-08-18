package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/persistence"
)

func TestRegisterEvents(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()

	provider.On("GetEventDBAL").Return(dbal)
	provider.On("GetTimeService").Return(ts)

	service := NewEventService(provider)

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  "uuid",
	}

	dbal.On("NewEventID").Once().Return("c70d3e0e-7e0b-4638-b320-ee11f5c61055")
	dbal.On("AddEvents", event).Once().Return(nil)
	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	err := service.RegisterEvents(event)
	assert.NoError(t, err)

	assert.NotEqual(t, "", event.Id, "RegisterEvents should assign an ID to the event")

	provider.AssertExpectations(t)
	dbal.AssertExpectations(t)
}
