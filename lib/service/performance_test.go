package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterPerformance(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	as := new(MockAlgoAPI)
	cts := new(MockComputeTaskAPI)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetAlgoService").Return(as)
	provider.On("GetPerformanceDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)
	service := NewPerformanceService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	metric := &asset.Algo{
		Key: "1da600d4-f8ad-45d7-92a0-7ff752a82275",
		Outputs: map[string]*asset.AlgoOutput{
			"auc": {
				Kind: asset.AssetKind_ASSET_PERFORMANCE,
			},
		}}
	as.On("GetAlgo", "1da600d4-f8ad-45d7-92a0-7ff752a82275").Return(metric, nil)

	task := &asset.ComputeTask{
		Key:      "taskTest",
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Worker:   "test",
		Category: asset.ComputeTaskCategory_TASK_TEST,
		Outputs: map[string]*asset.ComputeTaskOutput{
			"auc": {},
		},
		AlgoKey: metric.Key,
	}
	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(task, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "auc",
		MetricKey:                   "1da600d4-f8ad-45d7-92a0-7ff752a82275",
		PerformanceValue:            0.36492,
	}

	stored := &asset.Performance{
		ComputeTaskKey:   perf.ComputeTaskKey,
		MetricKey:        perf.MetricKey,
		PerformanceValue: perf.PerformanceValue,
		CreationDate:     timestamppb.New(time.Unix(1337, 0)),
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

	as.AssertExpectations(t)
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
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Worker:   "test",
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
	}, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "auc",
		MetricKey:                   "1da600d4-f8ad-45d7-92a0-7ff752a82275",
		PerformanceValue:            0.36492,
	}

	_, err := service.RegisterPerformance(perf, "test")
	assert.Error(t, err)

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterPerformanceInvalidOutput(t *testing.T) {
	as := new(MockAlgoAPI)
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetAlgoService").Return(as)
	service := NewPerformanceService(provider)

	metric := &asset.Algo{
		Key: "1da600d4-f8ad-45d7-92a0-7ff752a82275",
		Outputs: map[string]*asset.AlgoOutput{
			"auc": {
				Kind: asset.AssetKind_ASSET_UNKNOWN,
			},
		}}
	as.On("GetAlgo", "1da600d4-f8ad-45d7-92a0-7ff752a82275").Return(metric, nil)

	task := &asset.ComputeTask{
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Worker:   "test",
		Category: asset.ComputeTaskCategory_TASK_TEST,
		Outputs: map[string]*asset.ComputeTaskOutput{
			"auc": {},
		},
		AlgoKey: metric.Key,
	}
	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(task, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:              "08680966-97ae-4573-8b2d-6c4db2b3c532",
		ComputeTaskOutputIdentifier: "foo",
		MetricKey:                   metric.Key,
		PerformanceValue:            0.36492,
	}

	_, err := service.RegisterPerformance(perf, "test")
	assert.ErrorContains(t, err, "has no output named \"foo\"")
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrMissingTaskOutput, orcError.Kind)

	perf.ComputeTaskOutputIdentifier = "auc"
	_, err = service.RegisterPerformance(perf, "test")
	orcError = new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrIncompatibleKind, orcError.Kind)
	as.AssertExpectations(t)
	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}
