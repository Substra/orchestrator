package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
)

func TestRegisterProfilingStep(t *testing.T) {
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()

	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)

	service := NewProfilingService(provider)

	profilingStep := &asset.ProfilingStep{
		AssetKey: "b2e86700-6951-4cb6-9cac-f704cac548ce",
		Step:     "build_image",
		Duration: 10000,
	}
	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKind: asset.AssetKind_ASSET_PROFILING_STEP,
		AssetKey:  profilingStep.AssetKey,
		Asset:     &asset.Event_ProfilingStep{ProfilingStep: profilingStep},
	}
	es.On("RegisterEvents", e).Return(nil)

	err := service.RegisterProfilingStep(profilingStep)

	assert.NoError(t, err, "Registration of valid profiling step should not fail")

	ts.AssertExpectations(t)
}
