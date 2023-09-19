package service

import (
	"errors"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
)

func TestRegisterComputeTaskFailureReport(t *testing.T) {
	taskService := new(MockComputeTaskAPI)
	failureReportDBAL := new(persistence.MockFailureReportDBAL)
	eventService := new(MockEventAPI)
	timeService := new(MockTimeAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(taskService)
	provider.On("GetFailureReportDBAL").Return(failureReportDBAL)
	provider.On("GetEventService").Return(eventService)
	provider.On("GetTimeService").Return(timeService)
	service := NewFailureReportService(provider)

	transactionTime := time.Unix(1337, 0)
	timeService.On("GetTransactionTime").Once().Return(transactionTime)

	newFailureReport := &asset.NewFailureReport{
		AssetKey:  "08680966-97ae-4573-8b2d-6c4db2b3c532",
		AssetType: asset.FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
		ErrorType: asset.ErrorType_ERROR_TYPE_EXECUTION,
		LogsAddress: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	taskService.On("GetTask", newFailureReport.AssetKey).Once().Return(&asset.ComputeTask{
		Key:    newFailureReport.AssetKey,
		Status: asset.ComputeTaskStatus_STATUS_DOING,
		Worker: "test",
	}, nil)

	taskService.On("ApplyTaskAction", newFailureReport.AssetKey, asset.ComputeTaskAction_TASK_ACTION_FAILED, "failure report registered", "test").Once().Return(nil)

	storedFailureReport := &asset.FailureReport{
		AssetKey:     newFailureReport.AssetKey,
		AssetType:    asset.FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
		ErrorType:    newFailureReport.ErrorType,
		LogsAddress:  newFailureReport.LogsAddress,
		CreationDate: timestamppb.New(transactionTime),
		Owner:        "test",
	}
	failureReportDBAL.On("AddFailureReport", storedFailureReport).Once().Return(nil)

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  newFailureReport.AssetKey,
		AssetKind: asset.AssetKind_ASSET_FAILURE_REPORT,
		Asset:     &asset.Event_FailureReport{FailureReport: storedFailureReport},
	}
	eventService.On("RegisterEvents", event).Once().Return(nil)

	failureReport, err := service.RegisterFailureReport(newFailureReport, "test")
	assert.NoError(t, err)
	assert.Equal(t, failureReport.AssetKey, newFailureReport.AssetKey)

	taskService.AssertExpectations(t)
	failureReportDBAL.AssertExpectations(t)
	eventService.AssertExpectations(t)
	timeService.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterFunctionFailureReport(t *testing.T) {
	functionService := new(MockFunctionAPI)
	failureReportDBAL := new(persistence.MockFailureReportDBAL)
	eventService := new(MockEventAPI)
	timeService := new(MockTimeAPI)
	provider := newMockedProvider()
	provider.On("GetFunctionService").Return(functionService)
	provider.On("GetFailureReportDBAL").Return(failureReportDBAL)
	provider.On("GetEventService").Return(eventService)
	provider.On("GetTimeService").Return(timeService)
	service := NewFailureReportService(provider)

	transactionTime := time.Unix(1337, 0)
	timeService.On("GetTransactionTime").Once().Return(transactionTime)

	newFailureReport := &asset.NewFailureReport{
		AssetKey:  "08680966-97ae-4573-8b2d-6c4db2b3c532",
		AssetType: asset.FailedAssetKind_FAILED_ASSET_FUNCTION,
		ErrorType: asset.ErrorType_ERROR_TYPE_BUILD,
		LogsAddress: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	functionService.On("GetFunction", newFailureReport.AssetKey).Once().Return(&asset.Function{
		Key:    newFailureReport.AssetKey,
		Status: asset.FunctionStatus_FUNCTION_STATUS_BUILDING,
		Owner:  "test",
	}, nil)

	functionService.On("ApplyFunctionAction", newFailureReport.AssetKey, asset.FunctionAction_FUNCTION_ACTION_FAILED, "failure report registered", "test").Once().Return(nil)

	storedFailureReport := &asset.FailureReport{
		AssetKey:     newFailureReport.AssetKey,
		AssetType:    asset.FailedAssetKind_FAILED_ASSET_FUNCTION,
		ErrorType:    newFailureReport.ErrorType,
		LogsAddress:  newFailureReport.LogsAddress,
		CreationDate: timestamppb.New(transactionTime),
		Owner:        "test",
	}
	failureReportDBAL.On("AddFailureReport", storedFailureReport).Once().Return(nil)

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  newFailureReport.AssetKey,
		AssetKind: asset.AssetKind_ASSET_FAILURE_REPORT,
		Asset:     &asset.Event_FailureReport{FailureReport: storedFailureReport},
	}
	eventService.On("RegisterEvents", event).Once().Return(nil)

	failureReport, err := service.RegisterFailureReport(newFailureReport, "test")
	assert.NoError(t, err)
	assert.Equal(t, failureReport.AssetKey, newFailureReport.AssetKey)

	functionService.AssertExpectations(t)
	failureReportDBAL.AssertExpectations(t)
	eventService.AssertExpectations(t)
	timeService.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterFailureOnFailedTask(t *testing.T) {
	taskService := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(taskService)
	service := NewFailureReportService(provider)

	newFailureReport := &asset.NewFailureReport{
		AssetKey:  "08680966-97ae-4573-8b2d-6c4db2b3c532",
		AssetType: asset.FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
		ErrorType: asset.ErrorType_ERROR_TYPE_EXECUTION,
		LogsAddress: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	taskService.On("GetTask", newFailureReport.AssetKey).Once().Return(&asset.ComputeTask{
		Status: asset.ComputeTaskStatus_STATUS_FAILED,
		Worker: "test",
	}, nil)

	_, err := service.RegisterFailureReport(newFailureReport, "test")
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrBadRequest, orcError.Kind)

	taskService.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetFailure(t *testing.T) {
	dbal := new(persistence.MockFailureReportDBAL)
	provider := newMockedProvider()

	provider.On("GetFailureReportDBAL").Return(dbal)

	service := NewFailureReportService(provider)

	failureReport := &asset.FailureReport{
		AssetKey: "uuid",
	}

	dbal.On("GetFailureReport", failureReport.AssetKey).Once().Return(failureReport, nil)

	ret, err := service.GetFailureReport(failureReport.AssetKey)
	assert.NoError(t, err)
	assert.Equal(t, failureReport, ret)

	provider.AssertExpectations(t)
	dbal.AssertExpectations(t)
}
