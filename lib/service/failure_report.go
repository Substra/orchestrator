package service

import (
	"fmt"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FailureReportAPI interface {
	RegisterFailureReport(failure *asset.NewFailureReport, owner string) (*asset.FailureReport, error)
	GetFailureReport(assetKey string) (*asset.FailureReport, error)
}

type FailureReportServiceProvider interface {
	GetFailureReportService() FailureReportAPI
}

type FailureReportDependencyProvider interface {
	LoggerProvider
	persistence.FailureReportDBALProvider
	ComputeTaskServiceProvider
	FunctionServiceProvider
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
	s.GetLogger().Debug().Interface("failureReport", newFailureReport).Str("requester", requester).Msg("Registering new failure report")

	err := newFailureReport.Validate()
	if err != nil {
		return nil, errors.FromValidationError(asset.FailureReportKind, err)
	}
	switch newFailureReport.AssetType {
		case asset.FailedAssetKind_FAILED_ASSET_COMPUTE_TASK:
			err = s.processTaskFailure(newFailureReport.AssetKey, requester)
		case asset.FailedAssetKind_FAILED_ASSET_FUNCTION:
			err = s.processFunctionFailure(newFailureReport.AssetKey, requester)
		default:
			return nil, errors.NewBadRequest("can only register failure for asset_kind values function and compute task")
	}

	if err != nil {
		return nil, err
	}

	failureReport := &asset.FailureReport{
		AssetKey:     newFailureReport.AssetKey,
		AssetType:    newFailureReport.AssetType,
		ErrorType:    newFailureReport.ErrorType,
		LogsAddress:  newFailureReport.LogsAddress,
		CreationDate: timestamppb.New(s.GetTimeService().GetTransactionTime()),
		Owner:        requester,
	}

	err = s.GetFailureReportDBAL().AddFailureReport(failureReport)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  failureReport.AssetKey,
		AssetKind: asset.AssetKind_ASSET_FAILURE_REPORT,
		Asset:     &asset.Event_FailureReport{FailureReport: failureReport},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}

	return failureReport, nil
}

func (s *FailureReportService) GetFailureReport(assetKey string) (*asset.FailureReport, error) {
	s.GetLogger().Debug().Str("assetKey", assetKey).Msg("Get failure report")
	return s.GetFailureReportDBAL().GetFailureReport(assetKey)
}


func checkTaskPermissions(task *asset.ComputeTask, requester string) error {
	if task.Worker != requester {
		return errors.NewPermissionDenied(fmt.Sprintf("only %q worker can register failure report for compute task", task.Worker))
	}

	if task.Status != asset.ComputeTaskStatus_STATUS_DOING {
		return errors.NewBadRequest(fmt.Sprintf("cannot register failure report for task with status %q", task.Status.String()))
	}

	return nil
}

func (s *FailureReportService) processTaskFailure(taskKey string, requester string) error {
	task, err := s.GetComputeTaskService().GetTask(taskKey)
	if err != nil {
		return err
	}

	err = checkTaskPermissions(task, requester)
	if err != nil {
		return err
	}

	return s.GetComputeTaskService().ApplyTaskAction(taskKey, asset.ComputeTaskAction_TASK_ACTION_FAILED, "failure report registered", requester)
}

func checkFunctionPermissions(function *asset.Function, requester string) error {
	if function.Owner != requester {
		return errors.NewPermissionDenied(fmt.Sprintf("only %q owner can register failure report for function", function.Owner))
	}

	return nil
}

func (s *FailureReportService) processFunctionFailure(functionKey string, requester string) error {
	function, err := s.GetFunctionService().GetFunction(functionKey)
	if err != nil {
		return err
	}

	err = checkFunctionPermissions(function, requester)

	if err != nil {
		return err
	}

	return s.GetFunctionService().ApplyFunctionAction(functionKey, asset.FunctionAction_FUNCTION_ACTION_FAILED, "failure report registered", requester)
}