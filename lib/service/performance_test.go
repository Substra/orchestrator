package service

import (
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegisterPerformance(t *testing.T) {
	dbal := new(persistenceHelper.DBAL)
	cts := new(MockComputeTaskAPI)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetMetricDBAL").Return(dbal)
	provider.On("GetPerformanceDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	provider.On("GetTimeService").Return(ts)
	service := NewPerformanceService(provider)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	metric := &asset.Metric{Key: "1da600d4-f8ad-45d7-92a0-7ff752a82275"}
	dbal.On("MetricExists", "1da600d4-f8ad-45d7-92a0-7ff752a82275").Return(true, nil)

	task := &asset.ComputeTask{
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Worker:   "test",
		Category: asset.ComputeTaskCategory_TASK_TEST,
		Data: &asset.ComputeTask_Test{
			Test: &asset.TestTaskData{
				MetricKeys: []string{metric.Key},
			},
		},
	}
	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(task, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:   "08680966-97ae-4573-8b2d-6c4db2b3c532",
		MetricKey:        "1da600d4-f8ad-45d7-92a0-7ff752a82275",
		PerformanceValue: 0.36492,
	}

	stored := &asset.Performance{
		ComputeTaskKey:   perf.ComputeTaskKey,
		MetricKey:        perf.MetricKey,
		PerformanceValue: perf.PerformanceValue,
		CreationDate:     timestamppb.New(time.Unix(1337, 0)),
	}

	dbal.On("PerformanceExists", stored).Return(false, nil).Once()
	dbal.On("AddPerformance", stored).Once().Return(nil)
	dbal.On("CountComputeTaskPerformances", perf.ComputeTaskKey).Return(0, nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_PERFORMANCE,
		AssetKey:  stored.GetKey(),
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}
	es.On("RegisterEvents", event).Once().Return(nil)

	// Performance registration will initiate a task transition to done
	cts.On("applyTaskAction", task, transitionDone, mock.AnythingOfType("string")).Once().Return(nil)

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
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Worker:   "test",
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
	}, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:   "08680966-97ae-4573-8b2d-6c4db2b3c532",
		MetricKey:        "1da600d4-f8ad-45d7-92a0-7ff752a82275",
		PerformanceValue: 0.36492,
	}

	_, err := service.RegisterPerformance(perf, "test")
	assert.Error(t, err)

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}
