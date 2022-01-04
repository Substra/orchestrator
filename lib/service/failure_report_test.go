package service

import (
	"errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/stretchr/testify/assert"
)

func TestRegisterFailureReport(t *testing.T) {
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
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		LogsAddress: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	taskService.On("GetTask", newFailureReport.ComputeTaskKey).Once().Return(&asset.ComputeTask{
		Status: asset.ComputeTaskStatus_STATUS_FAILED,
		Worker: "test",
	}, nil)

	storedFailureReport := &asset.FailureReport{
		ComputeTaskKey: newFailureReport.ComputeTaskKey,
		LogsAddress:    newFailureReport.LogsAddress,
		CreationDate:   timestamppb.New(transactionTime),
	}
	failureReportDBAL.On("AddFailureReport", storedFailureReport).Once().Return(nil)

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  newFailureReport.ComputeTaskKey,
		AssetKind: asset.AssetKind_ASSET_FAILURE_REPORT,
	}
	eventService.On("RegisterEvents", event).Once().Return(nil)

	failureReport, err := service.RegisterFailureReport(newFailureReport, "test")
	assert.Equal(t, failureReport.ComputeTaskKey, newFailureReport.ComputeTaskKey)
	assert.NoError(t, err)

	taskService.AssertExpectations(t)
	failureReportDBAL.AssertExpectations(t)
	eventService.AssertExpectations(t)
	timeService.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestRegisterFailureOnNonFailedTask(t *testing.T) {
	cts := new(MockComputeTaskAPI)
	provider := newMockedProvider()
	provider.On("GetComputeTaskService").Return(cts)
	service := NewFailureReportService(provider)

	failureReport := &asset.NewFailureReport{
		ComputeTaskKey: "08680966-97ae-4573-8b2d-6c4db2b3c532",
		LogsAddress: &asset.Addressable{
			StorageAddress: "https://somewhere",
			Checksum:       "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
		},
	}

	cts.On("GetTask", failureReport.ComputeTaskKey).Once().Return(&asset.ComputeTask{
		Status: asset.ComputeTaskStatus_STATUS_DOING,
		Worker: "test",
	}, nil)

	_, err := service.RegisterFailureReport(failureReport, "test")
	assert.Error(t, err)
	orcError := new(orcerrors.OrcError)
	assert.True(t, errors.As(err, &orcError))
	assert.Equal(t, orcerrors.ErrBadRequest, orcError.Kind)

	cts.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestGetFailure(t *testing.T) {
	dbal := new(persistence.MockFailureReportDBAL)
	provider := newMockedProvider()

	provider.On("GetFailureReportDBAL").Return(dbal)

	service := NewFailureReportService(provider)

	failureReport := &asset.FailureReport{
		ComputeTaskKey: "uuid",
	}

	dbal.On("GetFailureReport", failureReport.ComputeTaskKey).Once().Return(failureReport, nil)

	ret, err := service.GetFailureReport(failureReport.ComputeTaskKey)
	assert.NoError(t, err)
	assert.Equal(t, failureReport, ret)

	provider.AssertExpectations(t)
	dbal.AssertExpectations(t)
}
