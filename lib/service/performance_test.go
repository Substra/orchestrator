package service

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	persistenceHelper "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterPerformance(t *testing.T) {
	dbal := new(persistenceHelper.MockDBAL)
	cts := new(MockComputeTaskService)
	es := new(MockEventService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	provider.On("GetPerformanceDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)
	service := NewPerformanceService(provider)

	task := &asset.ComputeTask{
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Worker:   "test",
		Category: asset.ComputeTaskCategory_TASK_TEST,
	}
	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(task, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:   "08680966-97ae-4573-8b2d-6c4db2b3c532",
		PerformanceValue: 0.36492,
	}

	stored := &asset.Performance{
		ComputeTaskKey:   perf.ComputeTaskKey,
		PerformanceValue: perf.PerformanceValue,
	}

	dbal.On("AddPerformance", stored).Once().Return(nil)

	event := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_PERFORMANCE,
		AssetKey:  perf.ComputeTaskKey,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
	}
	es.On("RegisterEvents", []*asset.Event{event}).Once().Return(nil)

	// Performance registration will initiate a task transition to done
	cts.On("applyTaskAction", task, transitionDone, mock.AnythingOfType("string")).Once().Return(nil)

	_, err := service.RegisterPerformance(perf, "test")
	assert.NoError(t, err)

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterPerformanceInvalidTask(t *testing.T) {
	cts := new(MockComputeTaskService)
	provider := new(MockServiceProvider)
	provider.On("GetComputeTaskService").Return(cts)
	service := NewPerformanceService(provider)

	cts.On("GetTask", "08680966-97ae-4573-8b2d-6c4db2b3c532").Return(&asset.ComputeTask{
		Status:   asset.ComputeTaskStatus_STATUS_DOING,
		Worker:   "test",
		Category: asset.ComputeTaskCategory_TASK_TRAIN,
	}, nil)

	perf := &asset.NewPerformance{
		ComputeTaskKey:   "08680966-97ae-4573-8b2d-6c4db2b3c532",
		PerformanceValue: 0.36492,
	}

	_, err := service.RegisterPerformance(perf, "test")
	assert.Error(t, err)

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}
