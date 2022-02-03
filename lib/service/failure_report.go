package service

import (
	"fmt"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FailureReportAPI interface {
	RegisterFailureReport(failure *asset.NewFailureReport, owner string) (*asset.FailureReport, error)
	GetFailureReport(computeTaskKey string) (*asset.FailureReport, error)
}

type FailureReportServiceProvider interface {
	GetFailureReportService() FailureReportAPI
}

type FailureReportDependencyProvider interface {
	LoggerProvider
	persistence.FailureReportDBALProvider
	ComputeTaskServiceProvider
	EventServiceProvider
	TimeServiceProvider
}

type FailureReportService struct {
	FailureReportDependencyProvider
}

func NewFailureReportService(provider FailureReportDependencyProvider) *FailureReportService {
	return &FailureReportService{provider}
}

func (s *FailureReportService) RegisterFailureReport(newFailureReport *asset.NewFailureReport, requester string) (*asset.FailureReport, error) {
	s.GetLogger().WithField("failureReport", newFailureReport).WithField("requester", requester).Debug("Registering new failure report")

	err := newFailureReport.Validate()
	if err != nil {
		return nil, errors.FromValidationError(asset.FailureReportKind, err)
	}

	task, err := s.GetComputeTaskService().GetTask(newFailureReport.ComputeTaskKey)
	if err != nil {
		return nil, err
	}

	if task.Worker != requester {
		return nil, errors.NewPermissionDenied(fmt.Sprintf("only %q worker can register failure report", task.Worker))
	}

	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		return nil, errors.NewBadRequest(fmt.Sprintf("cannot register failure report for task with status %q", task.Status.String()))
	}

	err = s.GetComputeTaskService().ApplyTaskAction(task.Key, asset.ComputeTaskAction_TASK_ACTION_FAILED, "failure report registered", requester)
	if err != nil {
		return nil, err
	}

	failureReport := &asset.FailureReport{
		ComputeTaskKey: newFailureReport.ComputeTaskKey,
		ErrorType:      newFailureReport.ErrorType,
		LogsAddress:    newFailureReport.LogsAddress,
		CreationDate:   timestamppb.New(s.GetTimeService().GetTransactionTime()),
	}

	err = s.GetFailureReportDBAL().AddFailureReport(failureReport)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  failureReport.ComputeTaskKey,
		AssetKind: asset.AssetKind_ASSET_FAILURE_REPORT,
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}

	return failureReport, nil
}

func (s *FailureReportService) GetFailureReport(computeTaskKey string) (*asset.FailureReport, error) {
	s.GetLogger().WithField("computeTaskKey", computeTaskKey).Debug("Get failure report")
	return s.GetFailureReportDBAL().GetFailureReport(computeTaskKey)
}
