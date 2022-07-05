package service

import (
	"fmt"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PerformanceAPI interface {
	RegisterPerformance(perf *asset.NewPerformance, requester string) (*asset.Performance, error)
	QueryPerformances(p *common.Pagination, filter *asset.PerformanceQueryFilter) ([]*asset.Performance, common.PaginationToken, error)
}

type PerformanceServiceProvider interface {
	GetPerformanceService() PerformanceAPI
}

type PerformanceDependencyProvider interface {
	LoggerProvider
	persistence.PerformanceDBALProvider
	ComputeTaskServiceProvider
	AlgoServiceProvider
	EventServiceProvider
	TimeServiceProvider
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
	s.GetLogger().WithField("taskKey", newPerf.ComputeTaskKey).WithField("metricKey", newPerf.MetricKey).WithField("requester", requester).Debug("Registering new performance")
	err := newPerf.Validate()
	if err != nil {
		return nil, errors.FromValidationError(asset.PerformanceKind, err)
	}

	task, err := s.GetComputeTaskService().GetTask(newPerf.ComputeTaskKey)
	if err != nil {
		return nil, err
	}

	if task.Worker != requester {
		return nil, errors.NewPermissionDenied(fmt.Sprintf("only %q worker can register performance", task.Worker))
	}

	if task.Category != asset.ComputeTaskCategory_TASK_TEST {
		return nil, errors.NewBadRequest("cannot register performance on non test task")
	}

	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot register performance for task with status %q", task.Status.String()))
	}

	algo, err := s.GetAlgoService().GetAlgo(newPerf.MetricKey)
	if err != nil {
		return nil, err
	}

	if algo.Category != asset.AlgoCategory_ALGO_METRIC {
		return nil, errors.NewInvalidAsset(fmt.Sprintf("MetricKey should point to an Algo with category %s", asset.AlgoCategory_ALGO_METRIC.String()))
	}

	perf := &asset.Performance{
		ComputeTaskKey:   newPerf.ComputeTaskKey,
		PerformanceValue: newPerf.PerformanceValue,
		CreationDate:     timestamppb.New(s.GetTimeService().GetTransactionTime()),
		MetricKey:        newPerf.MetricKey,
	}

	exists, err := s.GetPerformanceDBAL().PerformanceExists(perf)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.NewConflict(asset.PerformanceKind, perf.GetKey())
	}

	err = s.GetPerformanceDBAL().AddPerformance(perf)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  perf.GetKey(),
		AssetKind: asset.AssetKind_ASSET_PERFORMANCE,
		Asset:     &asset.Event_Performance{Performance: perf},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}

	// Mark the test task as done
	reason := fmt.Sprintf("Performances registered on %s by %s", task.Key, requester)
	err = s.GetComputeTaskService().applyTaskAction(task, transitionDone, reason)
	if err != nil {
		return nil, err
	}

	return perf, nil
}

func (s *PerformanceService) QueryPerformances(p *common.Pagination, filter *asset.PerformanceQueryFilter) ([]*asset.Performance, common.PaginationToken, error) {
	return s.GetPerformanceDBAL().QueryPerformances(p, filter)
}
