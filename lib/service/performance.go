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
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
)

type PerformanceAPI interface {
	RegisterPerformance(perf *asset.NewPerformance, requester string) (*asset.Performance, error)
	GetComputeTaskPerformance(key string) (*asset.Performance, error)
}

type PerformanceServiceProvider interface {
	GetPerformanceService() PerformanceAPI
}

type PerformanceDependencyProvider interface {
	persistence.PerformanceDBALProvider
	ComputeTaskServiceProvider
	EventServiceProvider
}

type PerformanceService struct {
	PerformanceDependencyProvider
}

func NewPerformanceService(provider PerformanceDependencyProvider) *PerformanceService {
	return &PerformanceService{provider}
}

// RegisterPerformance check asset validity and stores a new performance report for the given task.
// Note that the task key will also be the performance key (1:1 relationship).
func (s *PerformanceService) RegisterPerformance(newPerf *asset.NewPerformance, requester string) (*asset.Performance, error) {
	log.WithField("taskKey", newPerf.ComputeTaskKey).WithField("requester", requester).Debug("Registering new performance")
	err := newPerf.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrInvalidAsset, err.Error())
	}

	task, err := s.GetComputeTaskService().GetTask(newPerf.ComputeTaskKey)
	if err != nil {
		return nil, err
	}

	if task.Worker != requester {
		return nil, fmt.Errorf("%w: only \"%s\" worker can register performance", errors.ErrPermissionDenied, task.Worker)
	}

	if task.Category != asset.ComputeTaskCategory_TASK_TEST {
		return nil, fmt.Errorf("%w: cannot register performance on non test task", errors.ErrBadRequest)
	}

	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		return nil, fmt.Errorf("%w: cannot register performance for task with status \"%s\"", errors.ErrBadRequest, task.Status.String())
	}

	perf := &asset.Performance{
		ComputeTaskKey:   newPerf.ComputeTaskKey,
		PerformanceValue: newPerf.PerformanceValue,
	}

	err = s.GetPerformanceDBAL().AddPerformance(perf)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  perf.ComputeTaskKey,
		AssetKind: asset.AssetKind_ASSET_PERFORMANCE,
	}
	err = s.GetEventService().RegisterEvent(event)
	if err != nil {
		return nil, err
	}

	reason := fmt.Sprintf("Performance registered on %s by %s", task.Key, requester)
	err = s.GetComputeTaskService().applyTaskAction(task, transitionDone, reason)
	if err != nil {
		return nil, err
	}

	return perf, nil
}

func (s *PerformanceService) GetComputeTaskPerformance(key string) (*asset.Performance, error) {
	return s.GetPerformanceDBAL().GetComputeTaskPerformance(key)
}
