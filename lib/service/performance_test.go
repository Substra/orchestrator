package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterPerformance(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	cts := new(MockComputeTaskAPI)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetPerformanceDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)
	service := NewPerformanceService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	task := &asset.ComputeTask{
		Key:    "taskTest",
		Status: asset.ComputeTaskStatus_STATUS_DOING,
		Worker: "test",
		Outputs: map[string]*asset.ComputeTaskOutput{
			"auc": {},
		},
		FunctionKey: "1da600d4-f8ad-45d7-92a0-7ff752a82275",
	}
	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(task, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "auc",
		PerformanceValue:            0.36492,
	}

	stored := &asset.Performance{
		ComputeTaskKey:              perf.ComputeTaskKey,
		ComputeTaskOutputIdentifier: perf.ComputeTaskOutputIdentifier,
		PerformanceValue:            perf.PerformanceValue,
		CreationDate:                timestamppb.New(time.Unix(1337, 0)),
	}

	dbal.On("PerformanceExists", stored).Return(false, nil).Once()
	dbal.On("AddPerformance", stored, "auc").Once().Return(nil)

	output := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              perf.ComputeTaskKey,
		ComputeTaskOutputIdentifier: perf.ComputeTaskOutputIdentifier,
		AssetKind:                   asset.AssetKind_ASSET_PERFORMANCE,
		AssetKey:                    stored.GetKey(),
	}

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_PERFORMANCE,
		AssetKey:  stored.GetKey(),
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_Performance{Performance: stored},
	}

	cts.On("addComputeTaskOutputAsset", output).Once().Return(nil).NotBefore(
		es.On("RegisterEvents", event).Once().Return(nil),
	)

	_, err := service.RegisterPerformance(perf, "test")
	assert.NoError(t, err)

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestRegisterPerformanceInvalidTask(t *testing.T) {
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(cts)
	service := NewPerformanceService(provider)

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(&asset.ComputeTask{
		Status: asset.ComputeTaskStatus_STATUS_DOING,
		Worker: "test",
	}, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "auc",
		PerformanceValue:            0.36492,
	}

	_, err := service.RegisterPerformance(perf, "test")
	assert.Error(t, err)

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}
