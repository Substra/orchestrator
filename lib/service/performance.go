package service

import (
	"fmt"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
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
	s.GetLogger().Debug().
		Str("taskKey", newPerf.ComputeTaskKey).
		Str("metricKey", newPerf.MetricKey).
		Str("requester", requester).
		Msg("Registering new performance")
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

	_, err = s.GetAlgoService().GetAlgo(newPerf.MetricKey)
	if err != nil {
		return nil, err
	}

	if _, ok := task.Outputs[newPerf.ComputeTaskOutputIdentifier]; !ok {
		return nil, errors.NewMissingTaskOutput(task.Key, newPerf.ComputeTaskOutputIdentifier)
	}
	algoOutput, ok := task.Algo.Outputs[newPerf.ComputeTaskOutputIdentifier]
	if !ok {
		// This should never happen since task outputs are checked against algo on registration
		return nil, errors.NewInternal(fmt.Sprintf("missing algo output %q for task %q", newPerf.ComputeTaskOutputIdentifier, task.Key))
	}
	if algoOutput.Kind != asset.AssetKind_ASSET_PERFORMANCE {
		return nil, errors.NewIncompatibleTaskOutput(task.Key, newPerf.ComputeTaskOutputIdentifier, algoOutput.Kind.String(), asset.AssetKind_ASSET_PERFORMANCE.String())
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

	err = s.GetPerformanceDBAL().AddPerformance(perf, newPerf.ComputeTaskOutputIdentifier)
	if err != nil {
		return nil, err
	}

	outputAsset := &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              newPerf.ComputeTaskKey,
		ComputeTaskOutputIdentifier: newPerf.ComputeTaskOutputIdentifier,
		AssetKind:                   asset.AssetKind_ASSET_PERFORMANCE,
		AssetKey:                    perf.GetKey(),
	}
	err = s.GetComputeTaskService().addComputeTaskOutputAsset(outputAsset)
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

	return perf, nil
}

func (s *PerformanceService) QueryPerformances(p *common.Pagination, filter *asset.PerformanceQueryFilter) ([]*asset.Performance, common.PaginationToken, error) {
	return s.GetPerformanceDBAL().QueryPerformances(p, filter)
}
